package httphandler

import (
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/demeero/bricks/echobrick"
	"github.com/demeero/bricks/errbrick"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"

	"github.com/demeero/pocket-link/redirects/link"
)

func Setup(svcName string, e *echo.Echo, links *link.Links) {
	middlewares(svcName, e)
	e.Any("/*", redirect(links))
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
	})
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

func redirect(links *link.Links) echo.HandlerFunc {
	return func(c echo.Context) error {
		shortened := strings.Trim(c.Request().URL.Path, "/")
		u, err := links.Lookup(c.Request().Context(), shortened)
		if errors.Is(err, errbrick.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		return c.Redirect(http.StatusFound, u.String())
	}
}
