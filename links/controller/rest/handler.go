package rest

import (
	"errors"
	"net/http"

	"github.com/demeero/pocket-link/bricks/zaplogger"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.uber.org/zap"

	"github.com/demeero/pocket-link/links/service"
)

type createLink struct {
	Original string `json:"original,omitempty"`
}

func New(s *service.Service) http.Handler {
	e := echo.New()
	middlewares(e)
	g := e.Group("/api/v1/links")
	g.POST("/", create(s))
	return e
}

func middlewares(e *echo.Echo) {
	e.Pre(middleware.AddTrailingSlash())
	e.Use(otelecho.Middleware("links"))
	e.Use(zaplogger.EchoMiddleware(zap.L()))
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
