package otel

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func setEnv(t *testing.T, m map[string]string) {
	t.Helper()
	clearOTLPEnv(t)
	for k, v := range m {
		t.Setenv(k, v)
	}
}

func clearOTLPEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT",
		"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT",
		"OTEL_EXPORTER_OTLP_PROTOCOL",
		"OTEL_TRACES_EXPORTER",
		"OTEL_METRICS_EXPORTER",
		"OTEL_TRACES_SAMPLER",
		"OTEL_SERVICE_NAME",
		"OTEL_RESOURCE_ATTRIBUTES",
	} {
		t.Setenv(k, "")
	}
}

func freshRegistries(t *testing.T) *registries {
	t.Helper()
	reg := newRegistries()
	old := defaultRegistries
	defaultRegistries = reg
	t.Cleanup(func() { defaultRegistries = old })
	return reg
}

func resetProviders(t *testing.T) {
	t.Helper()
	_ = Shutdown(context.Background())
	t.Cleanup(func() { _ = Shutdown(context.Background()) })
}

func registerGRPC(built *bool) {
	RegisterTraceExporter(ProtocolGRPC, func(context.Context) (tracesdk.SpanExporter, error) {
		if built != nil {
			*built = true
		}
		return &stubSpanExporter{}, nil
	})
	RegisterMetricReader(ProtocolGRPC, func(context.Context) (metricsdk.Reader, error) {
		return metricsdk.NewManualReader(), nil
	})
}

func TestSettingsFrom(t *testing.T) {
	setEnv(t, nil)
	s := settingsFrom("svc")
	assert.Equal(t, "svc", s.service)
	assert.Equal(t, ProtocolHTTP, s.protocol, "the built-in protocol is the default")
	assert.False(t, s.hasTraceEndpoint)
	assert.False(t, s.hasMetricEndpoint)

	setEnv(t, map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318"})
	s = settingsFrom("")
	assert.True(t, s.hasTraceEndpoint, "the shared endpoint counts for traces")
	assert.True(t, s.hasMetricEndpoint, "the shared endpoint counts for metrics")

	for _, v := range []string{"", "http", "http/protobuf"} {
		setEnv(t, map[string]string{"OTEL_EXPORTER_OTLP_PROTOCOL": v})
		s = settingsFrom("svc")
		assert.Equal(t, ProtocolHTTP, s.protocol, v)
	}
}

func TestResolve(t *testing.T) {
	tests := []struct {
		name       string
		env        map[string]string
		wantTraces bool
		wantMetric bool
		err        string
	}{
		{name: "nothing configured records only"},
		{
			name:       "shared endpoint exports both",
			env:        map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318"},
			wantTraces: true, wantMetric: true,
		},
		{
			name:       "traces endpoint only",
			env:        map[string]string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://collector:4318"},
			wantTraces: true,
		},
		{
			name:       "metrics endpoint only",
			env:        map[string]string{"OTEL_EXPORTER_OTLP_METRICS_ENDPOINT": "http://collector:4318"},
			wantMetric: true,
		},
		{
			name: "exporter none wins over an endpoint",
			env: map[string]string{
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318",
				"OTEL_TRACES_EXPORTER":        "none",
				"OTEL_METRICS_EXPORTER":       "none",
			},
		},
		{
			name: "explicit otlp with an endpoint",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER":        "otlp",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318",
			},
			wantTraces: true, wantMetric: true,
		},
		{
			name: "explicit otlp without an endpoint",
			env:  map[string]string{"OTEL_TRACES_EXPORTER": "otlp"},
			err:  "traces: otlp requires an OTLP endpoint",
		},
		{
			name: "unregistered exporter value",
			env: map[string]string{
				"OTEL_TRACES_EXPORTER":        "console",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "http://collector:4318",
			},
			err: "unsupported exporter",
		},
		{
			name: "unknown protocol",
			env:  map[string]string{"OTEL_EXPORTER_OTLP_PROTOCOL": "bogus"},
			err:  "not a built-in or registered exporter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setEnv(t, tt.env)
			pl, err := resolve(settingsFrom("svc"), newRegistries())
			if tt.err != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.err)
				assert.Nil(t, pl, "a failed resolve plans nothing")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantTraces, pl.traceExporter != nil)
			assert.Equal(t, tt.wantMetric, pl.metricReader != nil)
		})
	}
}

func TestResolve_RegisteredName(t *testing.T) {
	reg := freshRegistries(t)
	RegisterTraceExporter("console", func(context.Context) (tracesdk.SpanExporter, error) {
		return &stubSpanExporter{}, nil
	})

	setEnv(t, map[string]string{"OTEL_TRACES_EXPORTER": "console"})
	pl, err := resolve(settingsFrom("svc"), reg)
	require.NoError(t, err)
	assert.NotNil(t, pl.traceExporter, "a registered name needs no endpoint")
	assert.Nil(t, pl.metricReader, "a traces registration leaves metrics alone")

	setEnv(t, map[string]string{"OTEL_METRICS_EXPORTER": "console"})
	_, err = resolve(settingsFrom("svc"), reg)
	require.Error(t, err, "a traces-only registration does not enable metrics")
	assert.Contains(t, err.Error(), "metrics:")
	assert.Contains(t, err.Error(), "unsupported exporter")

	RegisterMetricReader("console", func(context.Context) (metricsdk.Reader, error) {
		return metricsdk.NewManualReader(), nil
	})
	setEnv(t, map[string]string{"OTEL_METRICS_EXPORTER": "console"})
	pl, err = resolve(settingsFrom("svc"), reg)
	require.NoError(t, err)
	assert.NotNil(t, pl.metricReader)
}

func TestNewProviders_DoesNotTouchGlobals(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)

	before := otel.GetTracerProvider()
	p, err := newProviders(context.Background(), "svc")
	require.NoError(t, err)
	assert.Same(t, before, otel.GetTracerProvider(), "building does not install globals")
	assert.NotNil(t, p.tracer)

	require.NoError(t, p.shutdown(context.Background()))
	require.NoError(t, p.shutdown(context.Background()), "shutdown is idempotent")
}

func serviceName(res *resource.Resource) string {
	v, _ := res.Set().Value("service.name")
	return v.AsString()
}

func TestNewResource(t *testing.T) {
	clearOTLPEnv(t)
	res := newResource("test-service")
	assert.Equal(t, "test-service", serviceName(res))
}

func TestNewResource_Default(t *testing.T) {
	clearOTLPEnv(t)
	res := newResource("")
	assert.NotEmpty(t, serviceName(res))
}

func TestNewResource_ServiceNameWinsOverEnv(t *testing.T) {
	clearOTLPEnv(t)
	t.Setenv("OTEL_SERVICE_NAME", "env-service")
	res := newResource("code-service")
	assert.Equal(t, "code-service", serviceName(res))
}

func TestInit_RecordOnlyByDefault(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)

	require.NoError(t, Init("svc"))
	assert.NotNil(t, installed, "Init keeps the providers for the package Shutdown")

	_, span := Span(context.Background(), "op")
	span.End()
	assert.NotEmpty(t, span.SpanContext().TraceID().String())
}

func TestInit_SetsPropagator(t *testing.T) {
	clearOTLPEnv(t)
	resetProviders(t)

	require.NoError(t, Init("svc"))
	assert.Contains(t, otel.GetTextMapPropagator().Fields(), "traceparent")
}

func countingServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	hits := &atomic.Int32{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv, hits
}

func TestInit_ExportsToConfiguredEndpoint(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	srv, hits := countingServer(t)
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", srv.URL)

	require.NoError(t, Init("svc"))
	_, span := Span(context.Background(), "op")
	span.End()
	require.NoError(t, Shutdown(context.Background()))
	assert.Positive(t, hits.Load())
}

func TestInit_WithMetricEndpoint(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	srv, hits := countingServer(t)
	t.Setenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT", srv.URL)

	require.NoError(t, Init("svc"))
	_, isSDK := otel.GetMeterProvider().(*metricsdk.MeterProvider)
	assert.True(t, isSDK)

	counter, err := otel.Meter("test").Int64Counter("op.count")
	require.NoError(t, err)
	counter.Add(context.Background(), 1)

	require.NoError(t, Shutdown(context.Background()))
	assert.Positive(t, hits.Load())
}

func TestInit_ExporterNone(t *testing.T) {
	clearOTLPEnv(t)
	resetProviders(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4318")
	t.Setenv("OTEL_TRACES_EXPORTER", "none")
	t.Setenv("OTEL_METRICS_EXPORTER", "none")

	require.NoError(t, Init("svc"))

	_, span := Span(context.Background(), "op")
	span.End()
	assert.NotEmpty(t, span.SpanContext().TraceID().String())
}

func TestInit_SamplerFromEnv(t *testing.T) {
	clearOTLPEnv(t)
	resetProviders(t)
	t.Setenv("OTEL_TRACES_SAMPLER", "always_off")

	require.NoError(t, Init("svc"))

	_, span := Span(context.Background(), "op")
	assert.False(t, span.SpanContext().IsSampled())
	span.End()
}

type recordingExporter struct {
	names *[]string
}

func (e *recordingExporter) ExportSpans(_ context.Context, spans []tracesdk.ReadOnlySpan) error {
	for _, s := range spans {
		*e.names = append(*e.names, s.Name())
	}
	return nil
}

func (e *recordingExporter) Shutdown(context.Context) error { return nil }

func TestInit_RegisteredExporterName(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)

	var names []string
	RegisterTraceExporter("console", func(context.Context) (tracesdk.SpanExporter, error) {
		return &recordingExporter{names: &names}, nil
	})
	t.Setenv("OTEL_TRACES_EXPORTER", "console")

	require.NoError(t, Init("svc"))
	_, span := Span(context.Background(), "console-op")
	span.End()
	require.NoError(t, Shutdown(context.Background()))
	assert.Equal(t, []string{"console-op"}, names)
}

func TestInit_GRPCWithoutRegistration(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

	err := Init("svc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a built-in or registered exporter")
	assert.Nil(t, installed)
}

func TestInit_UnknownProtocolWithoutEndpoint(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "bogus")

	err := Init("svc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a built-in or registered exporter")
	assert.Nil(t, installed)
}

func TestInit_UnsupportedExporterValue(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "console")

	err := Init("svc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported exporter")
	assert.Nil(t, installed)
}

func TestInit_OtlpWithoutEndpoint(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_TRACES_EXPORTER", "otlp")

	err := Init("svc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires an OTLP endpoint")
	assert.Nil(t, installed)
}

func TestInit_TracesOnlyRegistration(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")
	RegisterTraceExporter(ProtocolGRPC, func(context.Context) (tracesdk.SpanExporter, error) {
		return &stubSpanExporter{}, nil
	})

	err := Init("svc")
	require.Error(t, err, "metrics never falls back to the built-in http")
	assert.Contains(t, err.Error(), "metrics:")
	assert.Contains(t, err.Error(), "RegisterMetricReader")
	assert.Nil(t, installed)
}

func TestInit_SwitchesProtocolByEnv(t *testing.T) {
	clearOTLPEnv(t)
	freshRegistries(t)
	resetProviders(t)
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
	t.Setenv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc")

	var built bool
	registerGRPC(&built)

	require.NoError(t, Init("svc"))
	assert.True(t, built)
	require.NoError(t, Shutdown(context.Background()))
}

func TestShutdown_WithoutInit(t *testing.T) {
	resetProviders(t)
	require.NoError(t, Shutdown(context.Background()))
}

func TestShutdown_MultipleCalls(t *testing.T) {
	clearOTLPEnv(t)
	resetProviders(t)

	require.NoError(t, Init("svc"))
	require.NoError(t, Shutdown(context.Background()))
	require.NoError(t, Shutdown(context.Background()))
}

func TestShutdown_ContextCancellation(t *testing.T) {
	clearOTLPEnv(t)
	resetProviders(t)
	require.NoError(t, Init("svc"))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_ = Shutdown(ctx)
}
