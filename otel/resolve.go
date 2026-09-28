package otel

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type plan struct {
	serviceName   string
	traceExporter factory[tracesdk.SpanExporter]
	metricReader  factory[metricsdk.Reader]
}

func resolve(s settings, reg *registries) (*plan, error) {
	if !reg.traces.has(s.protocol) && !reg.metrics.has(s.protocol) {
		return nil, unknownExporter(s.protocol)
	}

	traceExporter, err := resolveSignal(s.tracesExporter, s.protocol, s.hasTraceEndpoint, reg.traces)
	if err != nil {
		return nil, fmt.Errorf("traces: %w", err)
	}

	metricReader, err := resolveSignal(s.metricsExporter, s.protocol, s.hasMetricEndpoint, reg.metrics)
	if err != nil {
		return nil, fmt.Errorf("metrics: %w", err)
	}

	return &plan{serviceName: s.service, traceExporter: traceExporter, metricReader: metricReader}, nil
}

func resolveSignal[T any](value, protocol string, endpointSet bool, r *registry[T]) (factory[T], error) {
	key := value
	switch value {
	case "none":
		return nil, nil
	case "otlp":
		if !endpointSet {
			return nil, fmt.Errorf("otlp requires an OTLP endpoint")
		}
		key = protocol
	case "":
		if !endpointSet {
			return nil, nil
		}
		key = protocol
	}

	f, ok := r.get(key)
	switch {
	case ok:
		return f, nil
	case key == value:
		return nil, fmt.Errorf("unsupported exporter %q (want otlp, none, empty or a registered name)", value)
	default:
		return nil, unknownExporter(key)
	}
}

func (pl *plan) build(ctx context.Context) (*providers, error) {
	res := newResource(pl.serviceName)

	tracer, err := pl.newTracerProvider(ctx, res)
	if err != nil {
		return nil, err
	}

	meter, err := pl.newMeterProvider(ctx, res)
	if err != nil {
		return nil, errors.Join(err, tracer.Shutdown(ctx))
	}

	return &providers{tracer: tracer, meter: meter}, nil
}

func (pl *plan) newTracerProvider(ctx context.Context, res *resource.Resource) (*tracesdk.TracerProvider, error) {
	opts := []tracesdk.TracerProviderOption{tracesdk.WithResource(res)}
	if pl.traceExporter != nil {
		exp, err := pl.traceExporter(ctx)
		if err != nil {
			return nil, err
		}
		opts = append(opts, tracesdk.WithBatcher(exp))
	}
	return tracesdk.NewTracerProvider(opts...), nil
}

func (pl *plan) newMeterProvider(ctx context.Context, res *resource.Resource) (*metricsdk.MeterProvider, error) {
	if pl.metricReader == nil {
		return nil, nil
	}

	reader, err := pl.metricReader(ctx)
	if err != nil {
		return nil, err
	}
	return metricsdk.NewMeterProvider(metricsdk.WithResource(res), metricsdk.WithReader(reader)), nil
}

func newResource(serviceName string) *resource.Resource {
	def := resource.Default()
	if serviceName == "" {
		return def
	}
	res, err := resource.Merge(def, resource.NewSchemaless(attribute.String("service.name", serviceName)))
	if err != nil {
		otel.Handle(err)
		return def
	}
	return res
}

func unknownExporter(name string) error {
	return fmt.Errorf("%q is not a built-in or registered exporter; register it with otel.RegisterTraceExporter or otel.RegisterMetricReader", name)
}
