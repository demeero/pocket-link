package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"github.com/demeero/bricks/configbrick"
	"github.com/demeero/bricks/grpcbrick"
	"github.com/demeero/bricks/otelbrick"
	"github.com/demeero/bricks/slogbrick"
	"github.com/demeero/pocket-link/keygen/grpcsvc"
	"github.com/demeero/pocket-link/keygen/key"
	mongorepo "github.com/demeero/pocket-link/keygen/repository/mongo"
	redisrepo "github.com/demeero/pocket-link/keygen/repository/redis"
	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/go.mongodb.org/mongo-driver/mongo/otelmongo"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load environment variables from a `.env` file if one exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal("failed load .env file", err)
	}

	cfg := Config{}
	configbrick.LoadConfig(&cfg, os.Getenv("LOG_CONFIG") == "true")
	slogbrick.Configure(slogbrick.Config{
		Level:     cfg.Log.Level,
		AddSource: cfg.Log.AddSource,
		JSON:      cfg.Log.JSON,
	})

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
		SpanExclusions:        traceCfg.Exclusions,
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
		Exclusions:            meterCfg.Exclusions,
	})
	if err != nil {
		log.Fatalf("failed init metrics: %s", err)
	}

	usedRepo, err := createUsedKeysRepo(cfg)
	if err != nil {
		log.Fatal("failed create used keys repository", err)
	}
	unusedRepo := createUnusedKeysRepo(cfg.RedisUnusedKeys)

	defer cancel()

	genCfg := key.GeneratorConfig{
		PredefinedKeysCount: cfg.Generator.PredefinedKeysCount,
		Delay:               cfg.Generator.Delay,
		KeyLen:              cfg.Generator.KeyLen,
	}
	go key.Generate(ctx, genCfg, usedRepo, unusedRepo)

	grpcSrvShutdown := grpcServ(cfg.GRPC, key.New(cfg.Keys.TTL, usedRepo, unusedRepo))

	<-ctx.Done()
	slog.Info("shutting down")
	grpcSrvShutdown()
	if err := meterShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown meter provider", slog.Any("err", err))
	}
	if err := traceShutdown(context.Background()); err != nil {
		slog.Error("failed shutdown tracer provider", slog.Any("err", err))
	}
}

func grpcServ(cfg configbrick.GRPC, k *key.Keys) func() {
	interceptors := []grpc.UnaryServerInterceptor{
		grpcrecovery.UnaryServerInterceptor(),
		grpcbrick.SlogCtxUnaryServerInterceptor(true),
	}
	if cfg.AccessLog {
		interceptors = append(interceptors, grpcbrick.SlogUnaryServerInterceptor(slog.LevelDebug, func(ctx context.Context, _ interface{}, info *grpc.UnaryServerInfo) bool {
			return info.FullMethod == "/grpc.health.v1.Health/Check"
		}))
	}
	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()), grpc.ChainUnaryInterceptor(interceptors...))
	if cfg.EnableReflection {
		reflection.Register(grpcSrv)
	}
	pb.RegisterKeygenServiceServer(grpcSrv, grpcsvc.New(k))
	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcSrv, healthSrv)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal("failed listen GRPC port", err)
	}
	go func() {
		slog.Info("init grpc srv")
		healthSrv.SetServingStatus("keygen", grpc_health_v1.HealthCheckResponse_SERVING)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Fatal("failed serve GRPC", err)
		}
	}()
	return func() {
		healthSrv.Shutdown()
		grpcSrv.GracefulStop()
	}
}

func createUnusedKeysRepo(cfg configbrick.Redis) *redisrepo.UnusedKeys {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		DB:       cfg.DB,
		Password: cfg.Password,
	})
	if err := redisotel.InstrumentTracing(client); err != nil {
		slog.Error("failed instrument tracing to redis client for unused keys", slog.Any("err", err))
	}
	return redisrepo.NewUnusedKeys(client)
}

func createUsedKeysRepo(cfg Config) (key.UsedKeysRepository, error) {
	if cfg.UsedKeysRepositoryType == "" {
		cfg.UsedKeysRepositoryType = UsedKeysRepositoryTypeRedis
	}
	switch cfg.UsedKeysRepositoryType {
	case UsedKeysRepositoryTypeMongo:
		ctx, cancel := context.WithTimeout(context.Background(), cfg.MongoUsedKeys.InitialConnectTimeout)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoUsedKeys.URI).SetMonitor(otelmongo.NewMonitor()))
		if err != nil {
			return nil, fmt.Errorf("failed connect to MongoDB: %w", err)
		}
		return mongorepo.NewUsedKeys(client.Database("pocket-link"))
	case UsedKeysRepositoryTypeRedis:
		client := redis.NewClient(&redis.Options{
			Addr:     cfg.RedisUsedKeys.Addr,
			DB:       cfg.RedisUsedKeys.DB,
			Password: cfg.RedisUsedKeys.Password,
		})
		if err := redisotel.InstrumentTracing(client); err != nil {
			slog.Error("failed instrument tracing to redis client for used keys", slog.Any("err", err))
		}
		return redisrepo.NewUsedKeys(client), nil
	default:
		return nil, fmt.Errorf("unsupported used keys repository type: %s", cfg.UsedKeysRepositoryType)
	}
}
