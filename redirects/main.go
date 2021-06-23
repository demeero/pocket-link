package main

import (
	"log"
	"net/http"
	"strconv"

	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	"github.com/go-redis/redis/v8"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/demeero/pocket-link/redirects/config"
	"github.com/demeero/pocket-link/redirects/handler"
	"github.com/demeero/pocket-link/redirects/link"
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

	conn, err := grpc.Dial(cfg.Links.Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("error create grpc links connection", zap.Error(err))
	}

	l := link.New(linkpb.NewLinkServiceClient(conn), redis.NewClient(&redis.Options{Addr: cfg.RedisLRU.Addr}))
	httpServ(cfg.HTTP, l)
}

func httpServ(cfg config.HTTP, l *link.Links) {
	err := (&http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler.New(l),
	}).ListenAndServe()
	if err != nil {
		zap.L().Fatal("failed to listen HTTP: %v", zap.Error(err))
	}
}
