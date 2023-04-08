package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/demeero/pocket-link/bricks/trace"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/go-redis/redis/v8"
	"google.golang.org/grpc"

	"github.com/demeero/pocket-link/redirects/config"
	"github.com/demeero/pocket-link/redirects/handler"
	"github.com/demeero/pocket-link/redirects/link"
)

func main() {
	// Load environment variables from a `.env` file if one exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("failed load .env file")
	}

	cfg := config.New()
	configureLogger(cfg.Log)
	log.Debug().Any("value", cfg).Msg("parsed config")

	if err := trace.Init(context.Background(), "redirects", cfg.Telemetry.Collector.Addr); err != nil {
		log.Fatal().Err(err).Msg("error init tracing")
	}

	conn, err := grpc.Dial(cfg.Links.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		log.Fatal().Err(err).Msg("error create grpc links connection")
	}

	l := link.New(linkpb.NewLinkServiceClient(conn), redis.NewClient(&redis.Options{Addr: cfg.RedisLRU.Addr}))
	httpShutdown := httpSrv(cfg.HTTP, l)

	waitForShutdown(cfg.ShutdownTimeout, func(ctx context.Context) {
		log.Info().Msg("shutdown HTTP")
		httpShutdown(ctx)
		log.Info().Msg("shutdown links GRPC connection")
		if err := conn.Close(); err != nil {
			log.Error().Err(err).Msg("failed close grpc links connection")
		}
	})
}

func waitForShutdown(timeout time.Duration, shutdownFunc func(ctx context.Context)) {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
	<-sigint
	log.Info().Msg("start shutdown")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	shutdownFunc(ctx)
	log.Info().Msg("shutdown completed")
}

func httpSrv(cfg config.HTTP, l *link.Links) func(ctx context.Context) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(echolog.OFF)
	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout
	handler.Setup(e, l)
	go func() {
		log.Info().Msg("init HTTP srv")
		err := e.Start(fmt.Sprintf(":%d", cfg.Port))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("failed http serve")
		}
	}()
	return func(ctx context.Context) {
		if err := e.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed shutdown http srv")
		}
	}
}

func configureLogger(cfg config.Log) {
	if cfg.UnixTimestamp {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	if cfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
	if cfg.Caller {
		log.Logger = log.Logger.With().Caller().Logger()
	}
	zerolog.DefaultContextLogger = &log.Logger
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("failed parse log level")
	}
	zerolog.SetGlobalLevel(level)
}
