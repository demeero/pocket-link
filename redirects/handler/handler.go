package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/demeero/pocket-link/redirects/handler/middleware"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/demeero/pocket-link/redirects/link"
)

func Setup(e *echo.Echo, links *link.Links) {
	middlewares(e)
	e.Any("/*", redirect(links))
}

func middlewares(e *echo.Echo) {
	e.Pre(echomw.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(echomw.CORS())
	e.Use(otelecho.Middleware("redirects"))
	e.Use(middleware.Ctx())
	e.Use(middleware.Log())
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
