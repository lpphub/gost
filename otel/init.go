package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
)

// Init sets up the global OpenTelemetry providers; exporter/protocol/sampler
// config comes from the standard OTel environment variables.
func Init(opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	cfg.applyEnvOverrides()

	ctx := context.Background()
	res := newResource(cfg.serviceName)

	if err := initTraces(ctx, cfg, res); err != nil {
		return fmt.Errorf("init traces: %w", err)
	}

	if err := initMetrics(ctx, cfg, res); err != nil {
		_ = Shutdown(ctx)
		return fmt.Errorf("init metrics: %w", err)
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
	opts := []resource.Option{
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
	}
	if serviceName != "" {
		opts = append(opts, resource.WithAttributes(attribute.String("service.name", serviceName)))
	}

	res, err := resource.New(context.Background(), opts...)
	if err != nil {
		otel.Handle(err)
	}

	if serviceNameOf(res) == "" {
		if merged, mErr := resource.Merge(res, resource.NewSchemaless(attribute.String("service.name", "unknown_service"))); mErr == nil {
			res = merged
		}
	}
	return res
}

func serviceNameOf(res *resource.Resource) string {
	if res == nil {
		return ""
	}
	for _, kv := range res.Attributes() {
		if kv.Key == "service.name" {
			return kv.Value.AsString()
		}
	}
	return ""
}
