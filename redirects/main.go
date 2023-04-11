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
	"github.com/demeero/pocket-link/bricks/metric"
	"github.com/demeero/pocket-link/bricks/trace"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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
		log.Fatal().Err(err).Msg("failed init tracing")
	}

	metricShutdown, err := metric.Init(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("failed init metrics")
	}

	conn, err := grpc.Dial(cfg.Links.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		log.Fatal().Err(err).Msg("error create grpc links connection")
	}

	client := redis.NewClient(&redis.Options{Addr: cfg.RedisLRU.Addr, DB: int(cfg.RedisLRU.DB)})
	if err := redisotel.InstrumentTracing(client); err != nil {
		log.Error().Err(err).Msg("failed instrument redis client with tracing")
	}
	if err := redisotel.InstrumentMetrics(client); err != nil {
		log.Error().Err(err).Msg("failed instrument redis client with metrics")
	}
	l := link.New(linkpb.NewLinkServiceClient(conn), client)
	httpShutdown := httpSrv(cfg.HTTP, l)

	waitForShutdown(cfg.ShutdownTimeout, func(ctx context.Context) {
		log.Info().Msg("shutdown HTTP")
		httpShutdown(ctx)

		log.Info().Msg("shutdown tracing")
		if err := traceShutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed shutdown tracing")
		}

		log.Info().Msg("shutdown metrics")
		metricShutdown(ctx)

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
