package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/demeero/bricks/configbrick"
	"github.com/demeero/bricks/grpcbrick"
	"github.com/demeero/bricks/otelbrick"
	"github.com/demeero/bricks/slogbrick"
	"github.com/demeero/pocket-link/links/controller/rest"
	"github.com/demeero/pocket-link/links/controller/rpc"
	"github.com/demeero/pocket-link/links/repository"
	"github.com/demeero/pocket-link/links/service"
	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	"github.com/grafana/pyroscope-go"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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
	}

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
		SpanExclusions:        traceCfg.FormattedExclusions(),
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
		Exclusions:            meterCfg.FormattedExclusions(),
	})
	if err != nil {
		log.Fatalf("failed init metrics: %s", err)
	}

	mClient, mShutdown := mongoDB(cfg.Mongo)
	repo, err := repository.New(mClient.Database("pocket-link"))
	if err != nil {
		log.Fatalf("failed create repository: %s", err)
	}

	keygenClientConn, err := grpc.Dial(cfg.Keygen.Addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	if err != nil {
		log.Fatalf("failed create GRPC keygen connection: %s", err)
	}

	svc := service.New(repo, keygenpb.NewKeygenServiceClient(keygenClientConn))

	httpShutdown := httpSrv(cfg.ServiceName, cfg.HTTP, svc)
	grpcShutdown := grpcSrv(cfg.GRPC, svc)

	defer cancel()
	<-ctx.Done()
	slog.Info("shutting down")
	httpShutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	httpShutdown(httpShutdownCtx)
	grpcShutdown()
	mShutdown(context.Background())
	if err := meterShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown meter provider", slog.Any("err", err))
	}
	if err := traceShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown tracer provider", slog.Any("err", err))
	}
	if err := keygenClientConn.Close(); err != nil {
		slog.Error("failed shutdown grpc links connection", slog.Any("err", err))
	}
	if p != nil {
		if err := p.Stop(); err != nil {
			slog.Error("failed shutdown profiler", slog.Any("err", err))
		}
	}
}

func mongoDB(cfg configbrick.Mongo) (client *mongo.Client, shutdown func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.InitialConnectTimeout)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI).SetMonitor(otelmongo.NewMonitor()))
	cancel()
	if err != nil {
		log.Fatalf("failed connect to mongo: %s", err)
	}
	return client, func(ctx context.Context) {
		if err := client.Disconnect(ctx); err != nil {
			slog.Error("failed disconnect from mongo", slog.Any("err", err))
		}
	}
}

func httpSrv(svcName string, cfg configbrick.HTTP, s *service.Service) func(ctx context.Context) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(echolog.OFF)
	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout
	rest.Setup(svcName, e, s)
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

func grpcSrv(cfg configbrick.GRPC, s *service.Service) func() {
	interceptors := []grpc.UnaryServerInterceptor{
		grpcrecovery.UnaryServerInterceptor(),
		grpcbrick.SlogCtxUnaryServerInterceptor(true),
	}
	if cfg.AccessLog {
		interceptors = append(interceptors, grpcbrick.SlogUnaryServerInterceptor(slog.LevelDebug, func(ctx context.Context, _ interface{}, info *grpc.UnaryServerInfo) bool {
			return info.FullMethod == "/grpc.health.v1.Health/Check"
		}))
	}
	grpcServ := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(interceptors...))
	if cfg.EnableReflection {
		reflection.Register(grpcServ)
	}
	pb.RegisterLinkServiceServer(grpcServ, rpc.New(s))
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServ, healthSrv)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatalf("failed listen GRPC port: %s", err)
	}
	go func() {
		healthSrv.SetServingStatus("links", grpc_health_v1.HealthCheckResponse_SERVING)
		slog.Info("init grpc srv")
		if err := grpcServ.Serve(lis); err != nil {
			log.Fatalf("failed serve GRPC: %s", err)
		}
	}()
	return func() {
		healthSrv.Shutdown()
		grpcServ.GracefulStop()
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
