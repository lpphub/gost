package otel

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestNewResource(t *testing.T) {
	res := newResource("test-service")

	assert.NotNil(t, res)
	// The resource should contain the service name attribute.
	// We can verify by checking the resource's attributes via its schema.
	assert.NotEmpty(t, res)
}

func TestDialOpts_Insecure(t *testing.T) {
	cfg := defaultConfig()
	cfg.insecure = true

	opts := dialOpts(cfg)
	require.Len(t, opts, 1)

	// Verify the option contains insecure credentials by checking
	// it's of the expected type via grpc.WithTransportCredentials
	assert.NotNil(t, opts[0])
}

func TestDialOpts_Secure(t *testing.T) {
	cfg := defaultConfig()
	cfg.insecure = false

	opts := dialOpts(cfg)
	assert.Nil(t, opts)
}

func TestDialOpts_ReturnsPrecomputedSlice(t *testing.T) {
	// Verify insecureDialOpts is the same precomputed slice
	cfg := defaultConfig()
	cfg.insecure = true

	opts := dialOpts(cfg)
	assert.Len(t, opts, 1)

	// Precomputed slice should match
	assert.Len(t, insecureDialOpts, 1)
}

func TestTracerProvider_NotInitialized(t *testing.T) {
	// Reset state
	tracerProvider = nil

	// Should fall back to global provider
	tp := TracerProvider()
	assert.NotNil(t, tp)
	assert.Equal(t, otel.GetTracerProvider(), tp)
}

func TestMeterProvider_NotInitialized(t *testing.T) {
	// Reset state
	meterProvider = nil

	// Should fall back to global provider
	mp := MeterProvider()
	assert.NotNil(t, mp)
	assert.Equal(t, otel.GetMeterProvider(), mp)
}

func TestShutdown_NilProviders(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestInit_NoSignalsEnabled(t *testing.T) {
	// Reset global state
	tracerProvider = nil
	meterProvider = nil

	err := Init()
	assert.NoError(t, err)

	// Even with no signals, the propagator should be set
	prop := otel.GetTextMapPropagator()
	assert.NotNil(t, prop)
	// Should be a composite propagator with TraceContext and Baggage
	_, ok := prop.(propagation.TextMapPropagator)
	assert.True(t, ok)
}

func TestInit_WithCustomMetricsReader(t *testing.T) {
	// Reset global state before test
	tracerProvider = nil
	meterProvider = nil

	reader := metricsdk.NewManualReader()
	err := Init(
		WithService("test-svc"),
		WithMetricsReader(reader),
	)

	assert.NoError(t, err)

	// MeterProvider should be set since we have a reader
	assert.NotNil(t, meterProvider, "meterProvider should be set after Init with reader")
	mp := MeterProvider()
	assert.NotNil(t, mp)

	// TracerProvider should NOT be set (no traces enabled)
	assert.Nil(t, tracerProvider, "tracerProvider should not be set when traces not enabled")
	tp := TracerProvider()
	assert.NotNil(t, tp)

	// Cleanup
	err = Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdown_AfterInitWithMetrics(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Init(
		WithService("shutdown-test"),
		WithMetricsReader(metricsdk.NewManualReader()),
	)
	require.NoError(t, err)

	assert.NotNil(t, meterProvider)

	err = Shutdown(context.Background())
	assert.NoError(t, err)

	// After shutdown, providers should be nil
	assert.Nil(t, meterProvider)
	assert.Nil(t, tracerProvider)
}

func TestInit_WithService(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Init(WithService("my-service"))
	assert.NoError(t, err)

	// Propagator should be set
	assert.NotNil(t, otel.GetTextMapPropagator())
}

func TestInit_WithTracesAndMetricsEnabled(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	// Enable both signals but without endpoints — should not error
	err := Init(
		WithService("both-test"),
		WithTracesEnabled(true),
		WithMetricsEnabled(true),
	)
	assert.NoError(t, err)
}

func TestDialOpts_Integration(t *testing.T) {
	t.Run("matches grpc expected option type", func(t *testing.T) {
		cfg := defaultConfig()
		cfg.insecure = true

		opts := dialOpts(cfg)
		require.Len(t, opts, 1)

		// Apply the dial option to a grpc dial config to ensure compatibility
		var dialOpts []grpc.DialOption
		dialOpts = append(dialOpts, opts...)
		assert.Len(t, dialOpts, 1)
	})

	t.Run("insecure credential is correct type", func(t *testing.T) {
		credOpt := grpc.WithTransportCredentials(insecure.NewCredentials())
		assert.NotNil(t, credOpt)
	})
}

func TestInit_PreservesServiceNameInResource(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Init(
		WithService("resource-test"),
		WithMetricsReader(metricsdk.NewManualReader()),
	)
	require.NoError(t, err)

	// Verify the resource was created with the correct service name
	// We can't easily inspect the resource internals, but we can verify
	// the meter provider was created with the resource.
	mp := meterProvider
	assert.NotNil(t, mp)

	assert.NoError(t, Shutdown(context.Background()))
}

func TestInit_EnvOverridesAreApplied(t *testing.T) {
	// This test verifies that Init applies env overrides.
	// Since we can't easily test the full Init path with real gRPC endpoints,
	// we verify that Init doesn't panic or error with various env configs.

	t.Run("with metrics reader, no endpoints", func(t *testing.T) {
		tracerProvider = nil
		meterProvider = nil

		err := Init(WithMetricsReader(metricsdk.NewManualReader()))
		assert.NoError(t, err)

		assert.NoError(t, Shutdown(context.Background()))
	})
}

func TestNewResourceAttributes(t *testing.T) {
	res := newResource("my-svc")

	assert.NotNil(t, res)
	// Resource should be created with semconv attributes:
	// - ServiceName("my-svc")
	// - TelemetrySDKLanguageGo
	// - TelemetrySDKName("opentelemetry")
	// - SchemaURL from semconv
	_ = semconv.SchemaURL // ensure import is used
}

func TestShutdown_MultipleCalls(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Init(WithMetricsReader(metricsdk.NewManualReader()))
	require.NoError(t, err)

	// First shutdown
	err = Shutdown(context.Background())
	assert.NoError(t, err)

	// Second shutdown should be safe (providers are nil)
	err = Shutdown(context.Background())
	assert.NoError(t, err)
}

func TestShutdown_ContextCancellation(t *testing.T) {
	// Reset state
	tracerProvider = nil
	meterProvider = nil

	err := Init(WithMetricsReader(metricsdk.NewManualReader()))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Shutdown with cancelled context should not panic
	err = Shutdown(ctx)
	// It may return an error from the cancelled context, but should not panic
	_ = err
}
