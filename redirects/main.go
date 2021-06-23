package main

import (
	"log"
	"net/http"
	"strconv"

	linkpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/demeero/pocket-link/redirects/config"
	"github.com/demeero/pocket-link/redirects/handler"
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

	httpServ(cfg.HTTP, linkpb.NewLinkServiceClient(conn))
}

func httpServ(cfg config.HTTP, c linkpb.LinkServiceClient) {
	err := (&http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: handler.New(c),
	}).ListenAndServe()
	if err != nil {
		zap.L().Fatal("failed to listen HTTP: %v", zap.Error(err))
	}
}
