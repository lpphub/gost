package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type stubSpanExporter struct{}

func (s *stubSpanExporter) ExportSpans(context.Context, []tracesdk.ReadOnlySpan) error { return nil }
func (s *stubSpanExporter) Shutdown(context.Context) error                             { return nil }

func grpcEnv(endpointVar string) map[string]string {
	return map[string]string{
		"OTEL_EXPORTER_OTLP_PROTOCOL": "grpc",
		endpointVar:                   "http://localhost:4317",
	}
}

func TestRegistry_TraceExporter(t *testing.T) {
	reg := newRegistries()
	var built bool
	reg.traces.register(ProtocolGRPC, func(context.Context) (tracesdk.SpanExporter, error) {
		built = true
		return &stubSpanExporter{}, nil
	})

	setEnv(t, grpcEnv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"))
	pl, err := resolve(settingsFrom("svc"), reg)
	require.NoError(t, err)
	require.NotNil(t, pl.traceExporter)

	exp, err := pl.traceExporter(context.Background())
	require.NoError(t, err)
	assert.IsType(t, &stubSpanExporter{}, exp)
	assert.True(t, built)
}

func TestRegistry_MetricReader(t *testing.T) {
	reg := newRegistries()
	reg.metrics.register(ProtocolGRPC, func(context.Context) (metricsdk.Reader, error) {
		return metricsdk.NewManualReader(), nil
	})

	setEnv(t, grpcEnv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"))
	pl, err := resolve(settingsFrom("svc"), reg)
	require.NoError(t, err)
	require.NotNil(t, pl.metricReader)

	reader, err := pl.metricReader(context.Background())
	require.NoError(t, err)
	assert.IsType(t, &metricsdk.ManualReader{}, reader)
}

func TestRegistry_PerSignal(t *testing.T) {
	reg := newRegistries()
	reg.traces.register(ProtocolGRPC, func(context.Context) (tracesdk.SpanExporter, error) {
		return &stubSpanExporter{}, nil
	})

	setEnv(t, grpcEnv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	_, err := resolve(settingsFrom("svc"), reg)
	require.Error(t, err, "a traces-only registration leaves metrics unresolved")
	assert.Contains(t, err.Error(), "metrics:")
	assert.Contains(t, err.Error(), "RegisterMetricReader")
}

func TestRegistry_OverridesBuiltin(t *testing.T) {
	reg := newRegistries()
	reg.traces.register(ProtocolHTTP, func(context.Context) (tracesdk.SpanExporter, error) {
		return &stubSpanExporter{}, nil
	})
	reg.metrics.register(ProtocolHTTP, func(context.Context) (metricsdk.Reader, error) {
		return metricsdk.NewManualReader(), nil
	})

	setEnv(t, map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://localhost:4318"})
	pl, err := resolve(settingsFrom("svc"), reg)
	require.NoError(t, err)

	exp, err := pl.traceExporter(context.Background())
	require.NoError(t, err)
	assert.IsType(t, &stubSpanExporter{}, exp)

	reader, err := pl.metricReader(context.Background())
	require.NoError(t, err)
	assert.IsType(t, &metricsdk.ManualReader{}, reader)
}

func TestRegistry_UnknownExporter(t *testing.T) {
	setEnv(t, grpcEnv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	_, err := resolve(settingsFrom("svc"), newRegistries())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a built-in or registered exporter")
	assert.Contains(t, err.Error(), "RegisterTraceExporter")
}
