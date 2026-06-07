package otel

import (
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func Tracer(name string) trace.Tracer {
	return TracerProvider().Tracer(name)
}

func Meter(name string) otelmetric.Meter {
	return MeterProvider().Meter(name)
}
