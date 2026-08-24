package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	tracerProvider   *tracesdk.TracerProvider
	meterProvider    *metricsdk.MeterProvider
	insecureDialOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
)

func Init(opts ...Option) error {
	cfg := defaultConfig()

	for _, o := range opts {
		o(cfg)
	}
	cfg.applyEnvOverrides()

	ctx := context.Background()
	res := newResource(cfg.serviceName)

	if cfg.tracesEnabled {
		if err := initTraces(ctx, cfg, res); err != nil {
			return fmt.Errorf("init traces: %w", err)
		}
	}

	if cfg.metricsEnabled {
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
	var opts []tracesdk.TracerProviderOption
	opts = append(opts, tracesdk.WithResource(res))

	if endpoint := cfg.endpointFor(cfg.tracesEndpoint); endpoint != "" {
		exp, err := otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithDialOption(dialOpts(cfg)...),
		)
		if err != nil {
			return err
		}
		opts = append(opts, tracesdk.WithBatcher(exp))
	}

	tracerProvider = tracesdk.NewTracerProvider(opts...)
	otel.SetTracerProvider(tracerProvider)
	return nil
}

func initMetrics(ctx context.Context, cfg *config, res *resource.Resource) error {
	var opts []metricsdk.Option

	for _, r := range cfg.metricsReaders {
		opts = append(opts, metricsdk.WithReader(r))
	}

	endpoint := cfg.endpointFor(cfg.metricsEndpoint)
	if endpoint != "" {
		exp, err := otlpmetricgrpc.New(ctx,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithDialOption(dialOpts(cfg)...),
		)
		if err != nil {
			return err
		}
		opts = append(opts, metricsdk.WithReader(metricsdk.NewPeriodicReader(exp)))
	}

	if len(opts) == 0 {
		return nil
	}

	opts = append(opts, metricsdk.WithResource(res))
	meterProvider = metricsdk.NewMeterProvider(opts...)
	otel.SetMeterProvider(meterProvider)
	return nil
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
	return resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(serviceName),
		semconv.TelemetrySDKLanguageGo,
		semconv.TelemetrySDKName("opentelemetry"),
	)
}

func dialOpts(cfg *config) []grpc.DialOption {
	if !cfg.insecure {
		return nil
	}
	return insecureDialOpts
}
