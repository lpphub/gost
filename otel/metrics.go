package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	otelmetric "go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

var meterProvider *metricsdk.MeterProvider

func MeterProvider() otelmetric.MeterProvider {
	if meterProvider != nil {
		return meterProvider
	}
	return otel.GetMeterProvider()
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
