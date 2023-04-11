package rest

import (
	"errors"
	"net/http"
	"strings"

	"github.com/demeero/pocket-link/bricks/httpecho/middleware"
	"github.com/rs/zerolog/log"

	"github.com/demeero/pocket-link/links/service"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type createLink struct {
	Original string `json:"original,omitempty"`
}

func Setup(e *echo.Echo, s *service.Service) {
	middlewares(e)
	apiGroup := e.Group("/api")
	linksGroup := apiGroup.Group("/links")
	linksGroup.POST("", create(s))

	for _, r := range e.Routes() {
		if r == nil || !strings.Contains(r.Name, "links") {
			continue
		}
		log.Info().
			Str("method", r.Method).
			Str("path", r.Path).
			Str("name", r.Name).
			Msg("registered routes")
	}
}

func middlewares(e *echo.Echo) {
	e.Pre(echomw.RemoveTrailingSlash())
	e.Use(middleware.Recover())
	e.Use(echomw.CORS())
	e.Use(middleware.Meter())
	e.Use(otelecho.Middleware("links"))
	e.Use(middleware.Ctx())
	e.Use(middleware.Log())
}

func create(s *service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		cl := createLink{}
		if err := c.Bind(&cl); err != nil {
			return err
		}
		result, err := s.Create(c.Request().Context(), cl.Original)
		if errors.Is(err, service.ErrInvalid) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, result)
	}
}
