package otel

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func Tracer(name string) trace.Tracer {
	return TracerProvider().Tracer(name)
}

func Meter(name string) metric.Meter {
	return MeterProvider().Meter(name)
}

func Span(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return Tracer(name).Start(ctx, name, opts...)
}
