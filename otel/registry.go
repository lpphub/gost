package otel

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

const (
	ProtocolHTTP = "http/protobuf"
	ProtocolGRPC = "grpc"
)

type factory[T any] func(context.Context) (T, error)

type registry[T any] struct {
	mu        sync.RWMutex
	factories map[string]factory[T]
}

func newRegistry[T any](builtins map[string]factory[T]) *registry[T] {
	return &registry[T]{factories: builtins}
}

func (r *registry[T]) register(name string, f factory[T]) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.factories[name] = f
}

func (r *registry[T]) get(name string) (factory[T], bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	f, ok := r.factories[name]
	return f, ok
}

func (r *registry[T]) has(name string) bool {
	_, ok := r.get(name)
	return ok
}

type registries struct {
	traces  *registry[tracesdk.SpanExporter]
	metrics *registry[metricsdk.Reader]
}

var defaultRegistries = newRegistries()

func newRegistries() *registries {
	return &registries{
		traces:  newRegistry(map[string]factory[tracesdk.SpanExporter]{ProtocolHTTP: newHTTPTraceExporter}),
		metrics: newRegistry(map[string]factory[metricsdk.Reader]{ProtocolHTTP: newHTTPMetricReader}),
	}
}

func newHTTPTraceExporter(ctx context.Context) (tracesdk.SpanExporter, error) {
	return otlptracehttp.New(ctx)
}

func newHTTPMetricReader(ctx context.Context) (metricsdk.Reader, error) {
	exp, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, err
	}
	return metricsdk.NewPeriodicReader(exp), nil
}

// RegisterTraceExporter 注册一个名为 name 的 traces 导出器，之后可由环境变量选中。
func RegisterTraceExporter(name string, f func(context.Context) (tracesdk.SpanExporter, error)) {
	defaultRegistries.traces.register(name, f)
}

// RegisterMetricReader 注册一个名为 name 的 metrics 读取器，之后可由环境变量选中。
func RegisterMetricReader(name string, f func(context.Context) (metricsdk.Reader, error)) {
	defaultRegistries.metrics.register(name, f)
}
