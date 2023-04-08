package middleware

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Ctx is a middleware that adds values to request context.
func Ctx() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			reqLogger := log.With().Str("uri", req.RequestURI).
				Str("http_method", req.Method).
				Logger()
			ctx := reqLogger.WithContext(req.Context())
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
	}
}
