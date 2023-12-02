// nolint: cyclop // it's ok for main package to have big average complexity
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/demeero/bricks/configbrick"
	"github.com/demeero/bricks/echobrick"
	"github.com/demeero/bricks/otelbrick"
	"github.com/demeero/bricks/slogbrick"
	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/grafana/pyroscope-go"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/demeero/pocket-link/redirects/httphandler"
	"github.com/demeero/pocket-link/redirects/link"
)

func main() {
	// Load environment variables from a `.env` file if one exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal("failed load .env file", err)
	}

	cfg := config{}
	configbrick.LoadConfig(&cfg, os.Getenv("LOG_CONFIG") == "true")
	slogbrick.Configure(slogbrick.Config{
		Level:     cfg.Log.Level,
		AddSource: cfg.Log.AddSource,
		JSON:      cfg.Log.JSON,
	})

	stopProfiling := profiling(cfg)

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)

	traceCfg := cfg.OTEL.Trace
	traceShutdown, err := otelbrick.InitTrace(ctx, otelbrick.TraceConfig{
		ServiceName:           cfg.ServiceName,
		ServiceNamespace:      cfg.ServiceNamespace,
		DeploymentEnvironment: cfg.Env,
		OTELHTTPEndpoint:      traceCfg.Endpoint,
		OTELHTTPPathPrefix:    traceCfg.PathPrefix,
		Insecure:              traceCfg.Insecure,
		Headers:               traceCfg.BasicAuthHeader(),
	})
	if err != nil {
		log.Fatalf("failed init tracer: %s", err)
	}

	meterCfg := cfg.OTEL.Meter
	meterShutdown, err := otelbrick.InitMeter(ctx, otelbrick.MeterConfig{
		ServiceName:           cfg.ServiceName,
		ServiceNamespace:      cfg.ServiceNamespace,
		DeploymentEnvironment: cfg.Env,
		OTELHTTPEndpoint:      meterCfg.Endpoint,
		OTELHTTPPathPrefix:    meterCfg.PathPrefix,
		Insecure:              meterCfg.Insecure,
		RuntimeMetrics:        true,
		HostMetrics:           true,
		Headers:               meterCfg.BasicAuthHeader(),
	})
	if err != nil {
		log.Fatalf("failed init metrics: %s", err)
	}

	conn, err := grpc.Dial(cfg.Links.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("failed create grpc links connection: %s", err)
	}

	client := redis.NewClient(&redis.Options{Addr: cfg.RedisLRU.Addr, DB: cfg.RedisLRU.DB, Password: cfg.RedisLRU.Password})
	if err := redisotel.InstrumentTracing(client); err != nil {
		slog.Error("failed instrument redis client with tracing", slog.Any("err", err))
	}
	l := link.New(linkpb.NewLinkServiceClient(conn), client)
	httpShutdown := httpSrv(cfg.ServiceName, cfg.HTTP, l)

	defer cancel()
	<-ctx.Done()
	slog.Info("shutting down")
	httpShutdown(nil)
	if err := meterShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown meter provider", slog.Any("err", err))
	}
	if err := traceShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown tracer provider", slog.Any("err", err))
	}
	if err := client.Close(); err != nil {
		slog.Error("failed shutdown redis client", slog.Any("err", err))
	}
	if err := conn.Close(); err != nil {
		slog.Error("failed shutdown grpc links connection", slog.Any("err", err))
	}
	stopProfiling()
}

func httpSrv(svcName string, cfg configbrick.HTTP, l *link.Links) func(ctx context.Context) {
	e := echo.New()
	srv := e.Server
	srv.WriteTimeout = cfg.WriteTimeout
	srv.ReadTimeout = cfg.ReadTimeout
	srv.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = echobrick.ErrorHandler
	e.Logger.SetLevel(echolog.OFF)
	httphandler.Setup(svcName, e, l)
	go func() {
		slog.Info("init HTTP srv")
		err := e.Start(fmt.Sprintf(":%d", cfg.Port))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed http serve: %s", err)
		}
	}()
	return func(ctx context.Context) {
		if err := e.Shutdown(ctx); err != nil {
			slog.Error("failed shutdown http srv", slog.Any("err", err))
		}
	}
}

func profiling(cfg config) func() {
	if !cfg.Profiler.Enabled {
		return func() {}
	}
	p, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: fmt.Sprintf("%s.%s", cfg.ServiceNamespace, cfg.ServiceName),
		ServerAddress:   cfg.Profiler.ServerAddress,
		Logger:          nil,
		Tags:            map[string]string{"env": cfg.Env, "version": cfg.Version},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
	})
	if err != nil {
		slog.Error("failed start profiler", slog.Any("err", err))
		return func() {}
	}
	return func() {
		if err := p.Stop(); err != nil {
			slog.Error("failed shutdown profiler", slog.Any("err", err))
		}
	}
}
