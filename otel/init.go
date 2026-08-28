package otel

import (
	"context"
	"errors"
	"fmt"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var defaultService string

func Init(opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	cfg.applyEnvOverrides()

	defaultService = cfg.serviceName
	if defaultService == "" {
		defaultService = os.Getenv("OTEL_SERVICE_NAME")
	}

	ctx := context.Background()
	res := newResource(cfg.serviceName)

	if err := initTraces(ctx, cfg, res); err != nil {
		return fmt.Errorf("init traces: %w", err)
	}

	if cfg.enabledMetrics() {
		if err := initMetrics(ctx, cfg, res); err != nil {
			_ = Shutdown(ctx)
			return fmt.Errorf("init metrics: %w", err)
		}
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return nil
}

func Shutdown(ctx context.Context) error {
	var errs []error

	if tracerProvider != nil {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		tracerProvider = nil
	}

	if meterProvider != nil {
		if err := meterProvider.Shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
		meterProvider = nil
	}

	return errors.Join(errs...)
}

func newResource(serviceName string) *resource.Resource {
	attrs := []attribute.KeyValue{
		semconv.TelemetrySDKLanguageGo,
		semconv.TelemetrySDKName("opentelemetry"),
	}
	if serviceName != "" {
		attrs = append(attrs, semconv.ServiceName(serviceName))
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
	if err != nil {
		return resource.NewWithAttributes(semconv.SchemaURL, attrs...)
	}
	return res
}
