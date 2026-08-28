package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerProvider *tracesdk.TracerProvider
	meterProvider  *metricsdk.MeterProvider
	defaultService string
)

func Init(opts ...Option) error {
	cfg := defaultConfig()
	for _, o := range opts {
		o(cfg)
	}
	cfg.applyEnvOverrides()
	defaultService = cfg.serviceName

	ctx := context.Background()
	res := newResource(cfg.serviceName)

	// Traces are always recorded locally so trace_id reaches logs.
	if err := initTraces(ctx, cfg, res); err != nil {
		return fmt.Errorf("init traces: %w", err)
	}

	// Metrics are only built when a reader/endpoint is configured.
	if cfg.metricsEnabled() {
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

func initTraces(ctx context.Context, cfg *config, res *resource.Resource) error {
	opts := []tracesdk.TracerProviderOption{tracesdk.WithResource(res)}
	if cfg.sampler != nil {
		opts = append(opts, tracesdk.WithSampler(cfg.sampler))
	}

	if cfg.tracesExportEnabled() {
		if endpoint := cfg.endpointFor(cfg.tracesEndpoint); endpoint != "" {
			exp, err := newTraceExporter(ctx, cfg, endpoint)
			if err != nil {
				return err
			}
			opts = append(opts, tracesdk.WithBatcher(exp))
		}
	}

	tracerProvider = tracesdk.NewTracerProvider(opts...)
	otel.SetTracerProvider(tracerProvider)
	return nil
}

func newTraceExporter(ctx context.Context, cfg *config, endpoint string) (tracesdk.SpanExporter, error) {
	switch cfg.protocol {
	case ProtocolHTTP:
		return otlptracehttp.New(ctx,
			otlptracehttp.WithEndpoint(endpoint),
		)
	default:
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
		)
	}
}

func initMetrics(ctx context.Context, cfg *config, res *resource.Resource) error {
	var opts []metricsdk.Option

	for _, r := range cfg.metricsReaders {
		opts = append(opts, metricsdk.WithReader(r))
	}

	if endpoint := cfg.endpointFor(cfg.metricsEndpoint); endpoint != "" {
		reader, err := newMetricReader(ctx, cfg, endpoint)
		if err != nil {
			return err
		}
		opts = append(opts, metricsdk.WithReader(reader))
	}

	if len(opts) == 0 {
		return nil
	}

	opts = append(opts, metricsdk.WithResource(res))
	meterProvider = metricsdk.NewMeterProvider(opts...)
	otel.SetMeterProvider(meterProvider)
	return nil
}

func newMetricReader(ctx context.Context, cfg *config, endpoint string) (metricsdk.Reader, error) {
	switch cfg.protocol {
	case ProtocolHTTP:
		exp, err := otlpmetrichttp.New(ctx,
			otlpmetrichttp.WithEndpoint(endpoint),
		)
		if err != nil {
			return nil, err
		}
		return metricsdk.NewPeriodicReader(exp), nil
	default:
		exp, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpoint),
		)
		if err != nil {
			return nil, err
		}
		return metricsdk.NewPeriodicReader(exp), nil
	}
}

func TracerProvider() trace.TracerProvider {
	if tracerProvider != nil {
		return tracerProvider
	}
	return otel.GetTracerProvider()
}

func MeterProvider() otelmetric.MeterProvider {
	if meterProvider != nil {
		return meterProvider
	}
	return otel.GetMeterProvider()
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
