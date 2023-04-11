package metric

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
)

type Config struct {
	ServiceName       string `required:"true" split_words:"true" json:"service_name"`
	ServiceNamespace  string `required:"true" split_words:"true" default:"pocket-link" json:"service_namespace"`
	OTELCollectorAddr string `required:"true" split_words:"true" json:"otel_collector_addr"`
}

func Init(ctx context.Context) (func(context.Context), error) {
	var cfg Config
	if err := envconfig.Process("metric", &cfg); err != nil {
		return nil, fmt.Errorf("failed process config: %w", err)
	}
	log.Debug().Any("value", cfg).Msg("metric config")
	exp, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithInsecure(), otlpmetricgrpc.WithEndpoint(cfg.OTELCollectorAddr))
	if err != nil {
		return nil, fmt.Errorf("failed init metrics exporter: %w", err)
	}
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.ServiceNamespace),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed create resource: %w", err)
	}
	meterProvider := metric.NewMeterProvider(metric.WithResource(res), metric.WithReader(metric.NewPeriodicReader(exp)))
	global.SetMeterProvider(meterProvider)
	return func(context.Context) {
		if err := meterProvider.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed shutdown meter provider")
		}
	}, nil
}
