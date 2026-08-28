package otel

import (
	"os"

	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

type Protocol string

const (
	ProtocolGRPC Protocol = "grpc"
	ProtocolHTTP Protocol = "http/protobuf"
)

type config struct {
	serviceName     string
	protocol        Protocol
	sampler         tracesdk.Sampler
	metricsReaders  []metricsdk.Reader
	tracesExporter  string
	metricsExporter string
}

type Option func(*config)

func WithService(name string) Option {
	return func(c *config) { c.serviceName = name }
}

func WithMetricsReader(reader ...metricsdk.Reader) Option {
	return func(c *config) {
		c.metricsReaders = append(c.metricsReaders, reader...)
	}
}

func WithProtocol(p Protocol) Option {
	return func(c *config) { c.protocol = p }
}

func WithSampler(s tracesdk.Sampler) Option {
	return func(c *config) { c.sampler = s }
}

func defaultConfig() *config {
	return &config{protocol: ProtocolHTTP}
}

func (c *config) applyEnvOverrides() {
	if v := os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL"); v != "" {
		c.protocol = protocolFromEnv(v)
	}

	c.tracesExporter = os.Getenv("OTEL_TRACES_EXPORTER")
	c.metricsExporter = os.Getenv("OTEL_METRICS_EXPORTER")
}

func (c *config) enabledTracesExport() bool {
	return c.tracesExporter != "none"
}

func (c *config) enabledMetrics() bool {
	if c.metricsExporter == "none" {
		return false
	}
	return len(c.metricsReaders) > 0 || c.hasMetricEndpoint()
}

func (c *config) hasTraceEndpoint() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
}

func (c *config) hasMetricEndpoint() bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT") != ""
}

func protocolFromEnv(v string) Protocol {
	switch v {
	case "grpc":
		return ProtocolGRPC
	default:
		return ProtocolHTTP
	}
}
