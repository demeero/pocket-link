package trace

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlphttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

func Init(ctx context.Context, serviceName, otelCollectorAddr string) error {
	otlpExp, err := otlp.NewExporter(ctx, otlphttp.NewDriver(
		otlphttp.WithInsecure(),
		otlphttp.WithEndpoint(otelCollectorAddr),
	))
	if err != nil {
		return err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSyncer(otlpExp),
		sdktrace.WithResource(resource.NewWithAttributes(
			attribute.Key("service.name").String(serviceName),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return nil
}

func FromContext(ctx context.Context) (string, string) {
	if ctx == nil {
		return "", ""
	}
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	var spanID, traceID string
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
