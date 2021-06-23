package zaplogger

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func EchoMiddleware(logger *zap.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// spanID, traceID := trace.FromContext(c.Request().Context())
			spanID, traceID := "", ""
			tracedLogger := logger.With(zap.String("span", spanID), zap.String("trace", traceID))
			ctxWithTracedLogger := To(c.Request().Context(), tracedLogger)
			c.SetRequest(c.Request().WithContext(ctxWithTracedLogger))

			start := time.Now()

			tracedLogger.Debug("incoming HTTP req",
				zap.String("method", c.Request().Method),
				zap.String("uri", c.Request().RequestURI),
				zap.String("addr", c.Request().RemoteAddr))

			if err := next(c); err != nil {
				c.Error(err)
			}

			tracedLogger.Debug("finished handling HTTP req",
				zap.String("uri", c.Request().RequestURI),
				zap.Int("code", c.Response().Status),
				zap.String("duration", time.Since(start).String()))

			return nil
		}
	}
}

func GRPCUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// spanID, traceID := trace.FromContext(ctx)
		spanID, traceID := "", ""

		logger := From(ctx)
		logger = logger.With(zap.String("trace", traceID), zap.String("span", spanID))

		logger.Debug("incoming GRPC req", zap.String("method", info.FullMethod))

		startTime := time.Now()
		resp, err := handler(To(ctx, logger), req)
		logger = logger.With(zap.String("duration", time.Since(startTime).String()))

		if err != nil {
			logger = logger.With(zap.Error(err))
		}
		logger.Debug("finished handling GRPC req", zap.String("code", status.Code(err).String()))
		return resp, err
	}
}
