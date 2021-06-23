package main

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"

	"github.com/kelseyhightower/envconfig"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/demeero/pocket-link/links/config"
	"github.com/demeero/pocket-link/links/controller/rest"
	"github.com/demeero/pocket-link/links/repository"
	"github.com/demeero/pocket-link/links/service"
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

	mClient, err := mongoDB(cfg.Mongo)
	if err != nil {
		logger.Fatal("error init mongo client", zap.Error(err))
	}
	repo, err := repository.New(mClient.Database("pocket-link"))
	if err != nil {
		logger.Fatal("error create repository", zap.Error(err))
	}

	conn, err := grpc.Dial(cfg.Keygen.Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("error create grpc keygen connection", zap.Error(err))
	}

	svc := service.New(repo, keygenpb.NewKeygenServiceClient(conn))
	httpServ(cfg.HTTP, svc)
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
