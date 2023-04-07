package interceptor

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func LogUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		logger := log.Ctx(ctx).With().Str("grpc_method", info.FullMethod).Logger()
		ctx = logger.WithContext(ctx)
		logger.Info().Msg("incoming GRPC req")

		startTime := time.Now()
		resp, err := handler(ctx, req)

		if err != nil {
			logger = logger.With().Err(err).Logger()
		}
		logger.Info().
			Str("code", status.Code(err).String()).
			Str("duration", time.Since(startTime).String()).
			Msg("finished handling GRPC req")

		return resp, err
	}
}
