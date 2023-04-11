package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
)

// Meter is a middleware that records metrics for each request.
func Meter() echo.MiddlewareFunc {
	httpMeter := global.MeterProvider().Meter("meter-http-middleware")
	srvLatencyHist, err := httpMeter.Int64Histogram("http_server_latency")
	if err != nil {
		log.Error().Err(err).Msg("failed create server latency hist")
	}
	srvReqCounter, err := httpMeter.Int64Counter("http_server_request_count")
	if err != nil {
		log.Error().Err(err).Msg("failed create request count measure")
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now().UTC()

			err := next(c)
			if err != nil {
				c.Error(err)
			}

			attrs := []attribute.KeyValue{
				attribute.String("method", c.Request().Method),
				attribute.String("path", c.Path()),
				attribute.Int("status", c.Response().Status),
			}
			ctx := c.Request().Context()
			srvReqCounter.Add(ctx, 1, attrs...)
			srvLatencyHist.Record(ctx, time.Since(start).Milliseconds(), attrs...)
			return err
		}
	}
}
