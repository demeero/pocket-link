package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/demeero/pocket-link/bricks/trace"
	"github.com/demeero/pocket-link/bricks/zaplogger"
	"github.com/go-redis/redis/v8"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"

	"github.com/demeero/pocket-link/keygen/config"
	"github.com/demeero/pocket-link/keygen/controller/rpc"
	"github.com/demeero/pocket-link/keygen/key"
	mongorepo "github.com/demeero/pocket-link/keygen/repository/mongo"
	redisrepo "github.com/demeero/pocket-link/keygen/repository/redis"
)

func main() {
	logger, _, err := zaplogger.New(zaplogger.Config{Level: zap.DebugLevel})
	if err != nil {
		log.Fatal("failed init logger: ", err)
	}

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Fatal("failed process config: ", zap.Error(err))
	}
	logger.Sugar().Debugf("config: %+v", cfg)

	if err := trace.Init(context.Background(), "keygen", cfg.Telemetry.Collector.Addr); err != nil {
		logger.Fatal("error init tracing: ", zap.Error(err))
	}

	usedRepo, err := createUsedKeysRepo(cfg)
	if err != nil {
		logger.Fatal("failed create used keys repository", zap.Error(err))
	}
	unusedRepo := redisrepo.NewUnusedKeys(redis.NewClient(&redis.Options{
		Addr: cfg.RedisUnusedKeys.Addr,
		DB:   int(cfg.RedisUnusedKeys.DB),
	}))

	genCtx, genCancel := context.WithCancel(context.Background())
	go key.Generate(genCtx, cfg.Generator, usedRepo, unusedRepo)

	grpcSrvShutdown := grpcServ(cfg.GRPC, key.New(cfg.Keys, usedRepo, unusedRepo))

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	logger.Info("sig int received - start shutdown")
	genCancel()
	logger.Info("start grpc srv shutdown")
	grpcSrvShutdown()
	logger.Info("grpc srv finished")
	logger.Info("shutdown completed")

}

func grpcServ(cfg config.GRPC, k *key.Keys) func() {
	grpcServ := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcrecovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			zaplogger.GRPCUnaryServerInterceptor(),
		),
	)
	reflection.Register(grpcServ)
	pb.RegisterKeygenServiceServer(grpcServ, rpc.New(k))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		zap.L().Fatal("failed to listen GRPC port", zap.Error(err))
	}
	go func() {
		if err := grpcServ.Serve(lis); err != nil {
			zap.L().Fatal("failed serve GRPC", zap.Error(err))
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
