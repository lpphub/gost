package otel

import (
	"os"
	"strconv"

	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

// Protocol selects the OTLP transport.
type Protocol string

const (
	// ProtocolGRPC is the OTLP/gRPC transport (default).
	ProtocolGRPC Protocol = "grpc"
	// ProtocolHTTP is the OTLP/HTTP transport.
	ProtocolHTTP Protocol = "http/protobuf"
)

type config struct {
	serviceName     string
	tracesEndpoint  string
	metricsEndpoint string
	defaultEndpoint string
	protocol        Protocol
	sampler         tracesdk.Sampler
	tracesExporter  string
	metricsExporter string
	metricsReaders  []metricsdk.Reader
}

type Option func(*config)

func WithService(name string) Option {
	return func(c *config) { c.serviceName = name }
}

// WithEndpoint sets the shared OTLP endpoint for traces and metrics.
func WithEndpoint(endpoint string) Option {
	return func(c *config) { c.defaultEndpoint = endpoint }
}

func WithMetricsReader(reader ...metricsdk.Reader) Option {
	return func(c *config) {
		c.metricsReaders = append(c.metricsReaders, reader...)
	}
}

// WithProtocol selects the OTLP transport (gRPC or HTTP).
func WithProtocol(p Protocol) Option {
	return func(c *config) { c.protocol = p }
}

// WithSampler overrides the default AlwaysSample sampler.
func WithSampler(s tracesdk.Sampler) Option {
	return func(c *config) { c.sampler = s }
}

func defaultConfig() *config {
	return &config{
		protocol: ProtocolGRPC,
	}
}

func (c *config) endpointFor(specific string) string {
	if specific != "" {
		return specific
	}
	return c.defaultEndpoint
}

func (c *config) applyEnvOverrides() {
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		c.defaultEndpoint = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); v != "" {
		c.tracesEndpoint = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"); v != "" {
		c.metricsEndpoint = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); v != "" {
		c.protocol = protocolFromEnv(v)
	}
	if v := os.Getenv("OTEL_TRACES_SAMPLER"); v != "" {
		c.sampler = samplerFromEnv(v, os.Getenv("OTEL_TRACES_SAMPLER_ARG"))
	}

	c.tracesExporter = os.Getenv("OTEL_TRACES_EXPORTER")
	c.metricsExporter = os.Getenv("OTEL_METRICS_EXPORTER")
}

// tracesExportEnabled reports whether spans should be exported.
func (c *config) tracesExportEnabled() bool {
	return c.tracesExporter != "none"
}

// metricsEnabled reports whether a MeterProvider should be built.
func (c *config) metricsEnabled() bool {
	if c.metricsExporter == "none" {
		return false
	}
	return c.endpointFor(c.metricsEndpoint) != "" || len(c.metricsReaders) > 0
}

func protocolFromEnv(v string) Protocol {
	switch v {
	case "http/protobuf", "http/json":
		return ProtocolHTTP
	default:
		return ProtocolGRPC
	}
}

func samplerFromEnv(name, arg string) tracesdk.Sampler {
	switch name {
	case "always_off":
		return tracesdk.NeverSample()
	case "traceidratio":
		return tracesdk.TraceIDRatioBased(parseRatio(arg))
	case "parentbased_always_on":
		return tracesdk.ParentBased(tracesdk.AlwaysSample())
	case "parentbased_always_off":
		return tracesdk.ParentBased(tracesdk.NeverSample())
	case "parentbased_traceidratio":
		return tracesdk.ParentBased(tracesdk.TraceIDRatioBased(parseRatio(arg)))
	default:
		return tracesdk.AlwaysSample()
	}
}

func parseRatio(arg string) float64 {
	ratio, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return 0
	}
	if ratio < 0 {
		return 0
	}
	if ratio > 1 {
		return 1
	}
	return ratio
}
