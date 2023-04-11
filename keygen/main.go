package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/demeero/pocket-link/bricks"
	"github.com/demeero/pocket-link/bricks/trace"
	"github.com/demeero/pocket-link/keygen/controller/rpc/interceptor"
	"github.com/go-redis/redis/v8"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"

	"github.com/demeero/pocket-link/keygen/config"
	"github.com/demeero/pocket-link/keygen/controller/rpc"
	"github.com/demeero/pocket-link/keygen/key"
	mongorepo "github.com/demeero/pocket-link/keygen/repository/mongo"
	redisrepo "github.com/demeero/pocket-link/keygen/repository/redis"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from a `.env` file if one exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("failed load .env file")
	}

	cfg := config.New()
	bricks.ConfigureLogger(cfg.Log)
	log.Debug().Any("value", cfg).Msg("parsed config")

	traceShutdown, err := trace.Init(context.Background(), trace.Config{
		ServiceName:       "keygen",
		OTELCollectorAddr: cfg.Telemetry.Collector.Addr,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed init tracing")
	}

	usedRepo, err := createUsedKeysRepo(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed create used keys repository")
	}
	unusedRepo := redisrepo.NewUnusedKeys(redis.NewClient(&redis.Options{
		Addr: cfg.RedisUnusedKeys.Addr,
		DB:   int(cfg.RedisUnusedKeys.DB),
	}))

	genCtx, genCancel := context.WithCancel(context.Background())
	genCfg := key.GeneratorConfig{
		PredefinedKeysCount: cfg.Generator.PredefinedKeysCount,
		Delay:               cfg.Generator.Delay,
		KeyLen:              cfg.Generator.KeyLen,
	}
	go key.Generate(genCtx, genCfg, usedRepo, unusedRepo)

	grpcSrvShutdown := grpcServ(cfg.GRPC, key.New(cfg.Keys.TTL, usedRepo, unusedRepo))

	waitForShutdown(cfg.ShutdownTimeout, func(context.Context) {
		genCancel()

		log.Info().Msg("start trace shutdown")
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err = traceShutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed shutdown tracing")
		} else {
			log.Info().Msg("trace finished")
		}

		log.Info().Msg("start grpc srv shutdown")
		grpcSrvShutdown()
		log.Info().Msg("grpc srv finished")
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

func grpcServ(cfg config.GRPC, k *key.Keys) func() {
	grpcServ := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcrecovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			interceptor.LogUnaryServerInterceptor(),
		),
	)
	if cfg.EnableReflection {
		reflection.Register(grpcServ)
	}
	pb.RegisterKeygenServiceServer(grpcServ, rpc.New(k))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Fatal().Err(err).Msg("failed listen GRPC port")
	}
	go func() {
		log.Info().Msg("init grpc srv")
		if err := grpcServ.Serve(lis); err != nil {
			log.Fatal().Err(err).Msg("failed serve GRPC")
		}
	}()
	return func() {
		grpcServ.GracefulStop()
	}
}

func createUsedKeysRepo(cfg config.Config) (key.UsedKeysRepository, error) {
	if cfg.UsedKeysRepositoryType == "" {
		cfg.UsedKeysRepositoryType = config.UsedKeysRepositoryTypeRedis
	}
	switch cfg.UsedKeysRepositoryType {
	case config.UsedKeysRepositoryTypeMongo:
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoUsedKeys.URI))
		if err != nil {
			return nil, fmt.Errorf("failed connect to MongoDB: %w", err)
		}
		return mongorepo.NewUsedKeys(client.Database("pocket-link"))
	case config.UsedKeysRepositoryTypeRedis:
		return redisrepo.NewUsedKeys(redis.NewClient(&redis.Options{
			Addr: cfg.RedisUsedKeys.Addr,
			DB:   int(cfg.RedisUsedKeys.DB),
		})), nil
	default:
		return nil, fmt.Errorf("unsupported used keys repository type: %s", cfg.UsedKeysRepositoryType)
	}
}
