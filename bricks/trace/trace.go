package trace

import (
	"context"
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Config struct {
	Hostname          string `json:"hostname"`
	ServiceName       string `required:"true" split_words:"true" json:"service_name"`
	ServiceNamespace  string `required:"true" split_words:"true" default:"pocket-link" json:"service_namespace"`
	OTELCollectorAddr string `required:"true" split_words:"true" json:"otel_collector_addr"`
}

func Init(ctx context.Context) (func(context.Context) error, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed process config: %w", err)
	}
	log.Debug().Any("value", cfg).Msg("trace config")
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			// the service name used to display traces in backends
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceNamespace(cfg.ServiceNamespace),
			semconv.HostName(cfg.Hostname),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed create resource: %w", err)
	}

	conn, err := grpc.DialContext(ctx, cfg.OTELCollectorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed create gRPC connection to collector: %w", err)
	}

	// Set up a trace exporter
	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed create trace exporter: %w", err)
	}

	// Register the trace exporter with a TracerProvider, using a batch
	// span processor to aggregate spans before export.
	bsp := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tracerProvider)

	// set global propagator to tracecontext (the default is no-op).
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// Shutdown will flush any remaining spans and shut down the exporter.
	return tracerProvider.Shutdown, nil
}

func FromContext(ctx context.Context) (spanID, traceID string) {
	if ctx == nil {
		return "", ""
	}
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if !spanCtx.IsValid() {
		return "", ""
	}
	if spanCtx.SpanID().IsValid() {
		spanID = spanCtx.SpanID().String()
	}
	if spanCtx.TraceID().IsValid() {
		traceID = spanCtx.TraceID().String()
	}
	return spanID, traceID
}
