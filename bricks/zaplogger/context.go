package zaplogger

import (
	"context"

	"go.uber.org/zap"
)

type loggingContextType struct{}

var loggingContextKey = loggingContextType{}

// From Returns a logger set on the context, or the global zap.L() if none is found
func From(ctx context.Context) *zap.Logger {
	return FromOrDefault(ctx, zap.L())
}

// FromOrDefault returns a logger set on the context, or the caller specified default logger
func FromOrDefault(ctx context.Context, def *zap.Logger) *zap.Logger {
	logger, ok := ctx.Value(loggingContextKey).(*zap.Logger)
	if !ok || logger == nil {
		return def
	}
	return logger
}

func To(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, loggingContextKey, logger)
}
