package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Log is a middleware to provide logging for each request.
func Log() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			reqLogger := log.Ctx(req.Context())
			reqLogger.Info().Msg("received HTTP request")

			start := time.Now().UTC()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			res := c.Response()

			reqLogger.Info().
				Dur("req_duration", time.Since(start)).
				Int("resp_status", res.Status).
				Int64("resp_size", res.Size).
				Msg("completed handling HTTP request")

			return err
		}
	}
}
