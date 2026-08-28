package otel

import (
	"context"

	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Tracer returns a tracer named by name, falling back to the service name.
func Tracer(name string) trace.Tracer {
	if name == "" {
		name = defaultService
	}
	return TracerProvider().Tracer(name)
}

// Meter returns a meter named by name, falling back to the service name.
func Meter(name string) otelmetric.Meter {
	if name == "" {
		name = defaultService
	}
	return MeterProvider().Meter(name)
}

// Span starts a span named by name, falling back to the service name.
func Span(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if name == "" {
		name = defaultService
	}
	return Tracer(name).Start(ctx, name, opts...)
}
