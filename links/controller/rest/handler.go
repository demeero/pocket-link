package rest

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/demeero/bricks/echobrick"
	"github.com/demeero/bricks/errbrick"
	"github.com/demeero/pocket-link/links/service"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

type createLink struct {
	Original string `json:"original,omitempty"`
}

func Setup(svcName string, e *echo.Echo, s *service.Service) {
	middlewares(svcName, e)
	apiGroup := e.Group("/api")
	linksGroup := apiGroup.Group("/links")
	linksGroup.POST("", create(s))

	for _, r := range e.Routes() {
		if r == nil || !strings.Contains(r.Name, "links") {
			continue
		}
		slog.Info("registered routes",
			slog.String("method", r.Method), slog.String("path", r.Path), slog.String("name", r.Name))
	}
}

func middlewares(svcName string, e *echo.Echo) {
	meterMW, err := echobrick.OTELMeterMW(echobrick.OTELMeterMWConfig{
		Attrs: &echobrick.OTELMeterAttrsConfig{
			Method:     true,
			Path:       true,
			Status:     true,
			AttrsToCtx: true,
		},
		Metrics: &echobrick.OTELMeterMetricsConfig{
			ReqDuration: true,
			ReqCounter:  true,
			ReqSize:     true,
			RespSize:    true,
		},
	}, nil)
	if err != nil {
		log.Fatalf("failed create meter middleware: %s", err)
	}
	e.Pre(echomw.RemoveTrailingSlash())
	e.Use(echobrick.RecoverSlogMW())
	e.Use(echomw.CORS())
	e.Use(otelecho.Middleware(svcName))
	e.Use(meterMW)
	e.Use(echobrick.SlogCtxMW(echobrick.LogCtxMWConfig{Trace: true}))
	e.Use(echobrick.SlogLogMW(slog.LevelDebug, nil))
}

func create(s *service.Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		cl := createLink{}
		if err := c.Bind(&cl); err != nil {
			return err
		}
		result, err := s.Create(c.Request().Context(), cl.Original)
		if errors.Is(err, errbrick.ErrInvalidData) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, result)
	}
}
