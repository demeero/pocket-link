package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"

	"github.com/demeero/pocket-link/keygen/config"
	"github.com/demeero/pocket-link/keygen/controller/rpc"
	"github.com/demeero/pocket-link/keygen/key"
	redisrepo "github.com/demeero/pocket-link/keygen/repository/redis"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("error init logger: ", err)
	}
	zap.ReplaceGlobals(logger)

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Fatal("error process config: ", zap.Error(err))
	}
	logger.Sugar().Debugf("config: %+v", cfg)

	usedRepo := redisrepo.NewUsedKeys(redis.NewClient(&redis.Options{
		Addr: cfg.RedisUsedKeys.Addr,
		DB:   int(cfg.RedisUsedKeys.DB),
	}))
	unusedRepo := redisrepo.NewUnusedKeys(redis.NewClient(&redis.Options{
		Addr: cfg.RedisUnusedKeys.Addr,
		DB:   int(cfg.RedisUnusedKeys.DB),
	}))

	go key.Generate(context.Background(), cfg.Generator, usedRepo, unusedRepo)

	grpcServ(cfg.GRPC, key.New(cfg.Keys, usedRepo, unusedRepo))
}

func grpcServ(cfg config.GRPC, k *key.Keys) {
	grpcServ := grpc.NewServer()
	reflection.Register(grpcServ)
	pb.RegisterKeygenServiceServer(grpcServ, rpc.New(k))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		zap.L().Fatal("failed to listen GRPC port", zap.Error(err))
	}
	if err := grpcServ.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve GRPC", zap.Error(err))
	}
}
