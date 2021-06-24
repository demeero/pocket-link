package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/demeero/pocket-link/bricks/trace"
	"github.com/demeero/pocket-link/bricks/zaplogger"
	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/demeero/pocket-link/links/config"
	"github.com/demeero/pocket-link/links/controller/rest"
	"github.com/demeero/pocket-link/links/controller/rpc"
	"github.com/demeero/pocket-link/links/repository"
	"github.com/demeero/pocket-link/links/service"
)

func main() {
	logger, _, err := zaplogger.New(zaplogger.Config{Level: zap.DebugLevel})
	if err != nil {
		log.Fatal("error init logger: ", err)
	}

	var cfg config.Config
	if err := envconfig.Process("", &cfg); err != nil {
		logger.Fatal("error process config: ", zap.Error(err))
	}
	logger.Sugar().Debugf("config: %+v", cfg)

	if err := trace.Init(context.Background(), "links", cfg.Telemetry.Collector.Addr); err != nil {
		logger.Fatal("error init tracing: ", zap.Error(err))
	}

	mClient, err := mongoDB(cfg.Mongo)
	if err != nil {
		logger.Fatal("error init mongo client", zap.Error(err))
	}
	repo, err := repository.New(mClient.Database("pocket-link"))
	if err != nil {
		logger.Fatal("error create repository", zap.Error(err))
	}

	conn, err := grpc.Dial(cfg.Keygen.Addr, grpc.WithInsecure(), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		logger.Fatal("error create grpc keygen connection", zap.Error(err))
	}

	svc := service.New(repo, keygenpb.NewKeygenServiceClient(conn))

	go httpServ(cfg.HTTP, svc)
	grpcServ(cfg.GRPC, svc)
}

func mongoDB(cfg config.Mongo) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	zap.L().Info("connect and ping to mongo are successful", zap.String("URI", cfg.URI))
	return client, nil
}

func httpServ(cfg config.HTTP, s *service.Service) {
	err := (&http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: rest.New(s),
	}).ListenAndServe()
	if err != nil {
		zap.L().Fatal("failed to listen HTTP: %v", zap.Error(err))
	}
}

func grpcServ(cfg config.GRPC, s *service.Service) {
	grpcServ := grpc.NewServer(
		grpcmiddleware.WithUnaryServerChain(
			grpcrecovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			zaplogger.GRPCUnaryServerInterceptor(),
		),
	)
	reflection.Register(grpcServ)
	pb.RegisterLinkServiceServer(grpcServ, rpc.New(s))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		zap.L().Fatal("failed to listen GRPC port", zap.Error(err))
	}
	if err := grpcServ.Serve(lis); err != nil {
		zap.L().Fatal("failed to serve GRPC", zap.Error(err))
	}
}
