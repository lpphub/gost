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
	if !cfg.enabledMetrics() {
		return nil
	}

	var opts []metricsdk.Option

	for _, r := range cfg.metricsReaders {
		opts = append(opts, metricsdk.WithReader(r))
	}

	if cfg.hasMetricEndpoint() {
		reader, err := newMetricReader(ctx, cfg)
		if err != nil {
			return err
		}
		opts = append(opts, metricsdk.WithReader(reader))
	}

	opts = append(opts, metricsdk.WithResource(res))
	meterProvider = metricsdk.NewMeterProvider(opts...)
	otel.SetMeterProvider(meterProvider)
	return nil
}

func newMetricReader(ctx context.Context, cfg *config) (metricsdk.Reader, error) {
	var exp metricsdk.Exporter
	var err error

	switch cfg.protocol {
	case ProtocolHTTP:
		exp, err = otlpmetrichttp.New(ctx)
	default:
		exp, err = otlpmetricgrpc.New(ctx)
	}
	if err != nil {
		return nil, err
	}
	return metricsdk.NewPeriodicReader(exp), nil
}
