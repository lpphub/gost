package otel

import (
	"context"
	"errors"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type providers struct {
	tracer *tracesdk.TracerProvider
	meter  *metricsdk.MeterProvider
	once   sync.Once
}

func newProviders(ctx context.Context, service string) (*providers, error) {
	pl, err := resolve(settingsFrom(service), defaultRegistries)
	if err != nil {
		return nil, err
	}
	return pl.build(ctx)
}

func (p *providers) shutdown(ctx context.Context) error {
	if p == nil {
		return nil
	}
	var err error
	p.once.Do(func() {
		err = errors.Join(err, p.tracer.Shutdown(ctx))
		if p.meter != nil {
			err = errors.Join(err, p.meter.Shutdown(ctx))
		}
	})
	return err
}

var installed *providers

// Init 构建 provider 并安装为 otel 进程级全局（含 W3C tracecontext/baggage 传播器）。
func Init(service string) error {
	p, err := newProviders(context.Background(), service)
	if err != nil {
		return err
	}

	otel.SetTracerProvider(p.tracer)
	if p.meter == nil {
		otel.SetMeterProvider(noop.NewMeterProvider())
	} else {
		otel.SetMeterProvider(p.meter)
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	installed = p
	return nil
}

// Shutdown 关闭 Init 安装的 provider，幂等，未 Init 时返回 nil。
func Shutdown(ctx context.Context) error {
	p := installed
	installed = nil
	return p.shutdown(ctx)
}
