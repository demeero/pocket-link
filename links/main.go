package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/demeero/pocket-link/bricks/trace"
	"github.com/demeero/pocket-link/links/config"
	"github.com/demeero/pocket-link/links/controller/rest"
	"github.com/demeero/pocket-link/links/controller/rpc"
	"github.com/demeero/pocket-link/links/controller/rpc/interceptor"
	"github.com/demeero/pocket-link/links/repository"
	"github.com/demeero/pocket-link/links/service"
	"google.golang.org/grpc/credentials/insecure"

	keygenpb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/keygen/v1beta1"
	pb "github.com/demeero/pocket-link/proto/gen/go/pocketlink/link/v1beta1"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echolog "github.com/labstack/gommon/log"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Load environment variables from a `.env` file if one exists
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal().Err(err).Msg("failed load .env file")
	}

	cfg := config.New()
	configureLogger(cfg.Log)
	log.Debug().Any("value", cfg).Msg("parsed config")

	if err := trace.Init(context.Background(), "links", cfg.Telemetry.Collector.Addr); err != nil {
		log.Fatal().Err(err).Msg("failed init tracing")
	}

	mClient, mShutdown := mongoDB(cfg.Mongo)
	repo, err := repository.New(mClient.Database("pocket-link"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed create repository")
	}

	keygenClientConn, err := grpc.Dial(cfg.Keygen.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	if err != nil {
		log.Fatal().Err(err).Msg("failed create GRPC keygen connection")
	}

	svc := service.New(repo, keygenpb.NewKeygenServiceClient(keygenClientConn))

	httpShutdown := httpSrv(cfg.HTTP, svc)
	grpcShutdown := grpcSrv(cfg.GRPC, svc)

	waitForShutdown(cfg.ShutdownTimeout, func(ctx context.Context) {
		log.Info().Msg("shutdown HTTP")
		httpShutdown(ctx)
		log.Info().Msg("shutdown GRPC")
		if err := keygenClientConn.Close(); err != nil {
			log.Error().Err(err).Msg("failed close keygen client GRPC connection")
		}
		grpcShutdown()
		log.Info().Msg("shutdown MongoDB")
		mShutdown(ctx)
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

func mongoDB(cfg config.Mongo) (client *mongo.Client, shutdown func(ctx context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	cancel()
	if err != nil {
		log.Fatal().Err(err).Msg("failed connect to mongo")
	}
	return client, func(ctx context.Context) {
		if err := client.Disconnect(ctx); err != nil {
			log.Error().Err(err).Msg("failed disconnect from mongo")
		}
	}
}

func httpSrv(cfg config.HTTP, s *service.Service) func(ctx context.Context) {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(echolog.OFF)
	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.ReadHeaderTimeout = cfg.ReadHeaderTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout
	rest.Setup(e, s)
	go func() {
		log.Info().Msg("init HTTP srv")
		err := e.Start(fmt.Sprintf(":%d", cfg.Port))
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("failed http serve")
		}
	}()
	return func(ctx context.Context) {
		if err := e.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed shutdown http srv")
		}
	}
}

func grpcSrv(cfg config.GRPC, s *service.Service) func() {
	grpcServ := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpcrecovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			interceptor.LogUnaryServerInterceptor(),
		),
	)
	if cfg.EnableReflection {
		reflection.Register(grpcServ)
	}
	pb.RegisterLinkServiceServer(grpcServ, rpc.New(s))

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

func configureLogger(cfg config.Log) {
	if cfg.UnixTimestamp {
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	}
	if cfg.Pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
	if cfg.Caller {
		log.Logger = log.Logger.With().Caller().Logger()
	}
	zerolog.DefaultContextLogger = &log.Logger
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		log.Fatal().Err(err).Msg("failed parse log level")
	}
	zerolog.SetGlobalLevel(level)
}
