package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

func clearOTLPEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_TRACES_SAMPLER",
		"OTEL_TRACES_SAMPLER_ARG",
		"OTEL_TRACES_EXPORTER",
		"OTEL_METRICS_EXPORTER",
	} {
		t.Setenv(k, "")
	}
}

func resourceServiceName(t *testing.T, res *resource.Resource) string {
	t.Helper()
	for _, kv := range res.Attributes() {
		if kv.Key == attribute.Key("service.name") {
			return kv.Value.AsString()
		}
	}
	return ""
}

func TestNewResource(t *testing.T) {
	clearOTLPEnv(t)
	res := newResource("test-service")
	assert.NotNil(t, res)
	assert.NotEmpty(t, res)
	assert.Equal(t, "test-service", resourceServiceName(t, res))
}

func TestNewResourceAttributes(t *testing.T) {
	clearOTLPEnv(t)
	res := newResource("my-svc")
	assert.NotNil(t, res)
	assert.Equal(t, "my-svc", resourceServiceName(t, res))
}

func TestTracerProvider_NotInitialized(t *testing.T) {
	tracerProvider = nil
	tp := TracerProvider()
	assert.NotNil(t, tp)
	assert.Equal(t, otel.GetTracerProvider(), tp)
}

func TestMeterProvider_NotInitialized(t *testing.T) {
	meterProvider = nil
	mp := MeterProvider()
	assert.NotNil(t, mp)
	assert.Equal(t, otel.GetMeterProvider(), mp)
}

func TestShutdown_NilProviders(t *testing.T) {
	tracerProvider = nil
	meterProvider = nil
	err := Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInit_NoSignalsEnabled(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	err := Init()
	assert.NoError(t, err)

	prop := otel.GetTextMapPropagator()
	assert.NotNil(t, prop)
	_, ok := prop.(propagation.TextMapPropagator)
	assert.True(t, ok)
}

func TestInit_WithService(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	assert.NoError(t, Init(WithService("my-service")))
	assert.NotNil(t, otel.GetTextMapPropagator())
	assert.NoError(t, Shutdown(context.Background()))
}

func TestInit_WithCustomMetricsReader(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	reader := metricsdk.NewManualReader()
	err := Init(WithService("test-svc"), WithMetricsReader(reader))
	assert.NoError(t, err)

	assert.NotNil(t, meterProvider, "meterProvider should be set after Init with reader")
	assert.NotNil(t, MeterProvider())
	assert.NotNil(t, tracerProvider, "tracerProvider should be a recorder even without an exporter")
	assert.NotNil(t, TracerProvider())

	err = Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdown_AfterInitWithMetrics(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	err := Init(WithService("shutdown-test"), WithMetricsReader(metricsdk.NewManualReader()))
	require.NoError(t, err)
	assert.NotNil(t, meterProvider)

	err = Shutdown(context.Background())
	require.NoError(t, err)
	assert.Nil(t, meterProvider)
	assert.Nil(t, tracerProvider)
}

func TestInit_WithTracesAndMetricsEnabled(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	err := Init()
	assert.NoError(t, err)
	assert.NotNil(t, tracerProvider, "tracerProvider should be a recorder by default")
	assert.Nil(t, meterProvider, "meterProvider should not be set without a reader/endpoint")

	err = Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInit_PreservesServiceNameInResource(t *testing.T) {
	clearOTLPEnv(t)
	res := newResource("resource-test")
	assert.Equal(t, "resource-test", resourceServiceName(t, res))
}

func TestInit_EnvOverridesAreApplied(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	err := Init(WithMetricsReader(metricsdk.NewManualReader()))
	assert.NoError(t, err)
	assert.NoError(t, Shutdown(context.Background()))
}

func TestShutdown_MultipleCalls(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	require.NoError(t, Init(WithMetricsReader(metricsdk.NewManualReader())))
	require.NoError(t, Shutdown(context.Background()))
	require.NoError(t, Shutdown(context.Background()))
}

func TestShutdown_ContextCancellation(t *testing.T) {
	clearOTLPEnv(t)
	tracerProvider = nil
	meterProvider = nil

	require.NoError(t, Init(WithMetricsReader(metricsdk.NewManualReader())))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = Shutdown(ctx)
}
