package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// Recover recovers from panics and logs the stack trace.
// It returns a 500 status code.
func Recover() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if err := recover(); err != nil {
					c.Response().WriteHeader(http.StatusInternalServerError)
					log.Ctx(c.Request().Context()).Error().Stack().
						Msgf("handler panicked: %+v", err)
				}
			}()
			return next(c)
		}
	}
}
