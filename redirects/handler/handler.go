package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/demeero/pocket-link/bricks/zaplogger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	"github.com/demeero/pocket-link/redirects/link"
)

func New(links *link.Links) http.Handler {
	e := echo.New()
	middlewares(e)
	e.Any("/*", redirect(links))
	return e
}

func middlewares(e *echo.Echo) {
	e.Pre(middleware.AddTrailingSlash())
	e.Use(otelecho.Middleware("redirects"))
	e.Use(zaplogger.EchoMiddleware(zap.L()))
}

func redirect(links *link.Links) echo.HandlerFunc {
	return func(c echo.Context) error {
		shortened := strings.Trim(c.Request().URL.Path, "/")
		u, err := links.Lookup(c.Request().Context(), shortened)
		if errors.Is(err, link.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.Redirect(http.StatusFound, u.String())
	}
}
