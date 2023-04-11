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

	"github.com/demeero/pocket-link/bricks"
	"github.com/demeero/pocket-link/bricks/trace"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
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
	bricks.ConfigureLogger(cfg.Log)
	log.Debug().Any("value", cfg).Msg("parsed config")

	traceShutdown, err := trace.Init(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("error init tracing")
	}

	conn, err := grpc.Dial(cfg.Links.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		log.Fatal().Err(err).Msg("error create grpc links connection")
	}

	l := link.New(linkpb.NewLinkServiceClient(conn), redis.NewClient(&redis.Options{Addr: cfg.RedisLRU.Addr, DB: int(cfg.RedisLRU.DB)}))
	httpShutdown := httpSrv(cfg.HTTP, l)

	waitForShutdown(cfg.ShutdownTimeout, func(ctx context.Context) {
		log.Info().Msg("shutdown tracing")
		if err := traceShutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed shutdown tracing")
		} else {
			log.Info().Msg("tracing shutdown completed")
		}

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
