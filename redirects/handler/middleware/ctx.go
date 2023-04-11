package middleware

import (
	"github.com/demeero/pocket-link/bricks/trace"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Ctx is a middleware that adds values to request context.
func Ctx() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			spanID, traceID := trace.FromContext(req.Context())
			reqLogger := log.With().Str("uri", req.RequestURI).
				Str("span_id", spanID).
				Str("trace_id", traceID).
				Str("http_method", req.Method).
				Logger()
			ctx := reqLogger.WithContext(req.Context())
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}
