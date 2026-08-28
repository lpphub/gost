package otel

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var tracerProvider *tracesdk.TracerProvider

func TracerProvider() trace.TracerProvider {
	if tracerProvider != nil {
		return tracerProvider
	}
	return otel.GetTracerProvider()
}

func initTraces(ctx context.Context, cfg *config, res *resource.Resource) error {
	opts := []tracesdk.TracerProviderOption{tracesdk.WithResource(res)}
	if cfg.sampler != nil {
		opts = append(opts, tracesdk.WithSampler(cfg.sampler))
	}

	if cfg.enabledTracesExport() && cfg.hasTraceEndpoint() {
		exp, err := newTraceExporter(ctx, cfg)
		if err != nil {
			return err
		}
		opts = append(opts, tracesdk.WithBatcher(exp))
	}

	tracerProvider = tracesdk.NewTracerProvider(opts...)
	otel.SetTracerProvider(tracerProvider)
	return nil
}

func newTraceExporter(ctx context.Context, cfg *config) (tracesdk.SpanExporter, error) {
	switch cfg.protocol {
	case ProtocolHTTP:
		return otlptracehttp.New(ctx)
	default:
		return otlptracegrpc.New(ctx)
	}
}
