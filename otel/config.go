package otel

import (
	"os"

	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

type config struct {
	serviceName     string
	tracesEnabled   bool
	metricsEnabled  bool
	tracesEndpoint  string
	metricsEndpoint string
	defaultEndpoint string
	insecure        bool
	metricsReaders  []metricsdk.Reader
}

type Option func(*config)

func WithService(name string) Option {
	return func(c *config) { c.serviceName = name }
}

func WithOTLPEndpoint(endpoint string) Option {
	return func(c *config) { c.defaultEndpoint = endpoint }
}

func WithTracesEndpoint(endpoint string) Option {
	return func(c *config) { c.tracesEndpoint = endpoint }
}

func WithMetricsEndpoint(endpoint string) Option {
	return func(c *config) { c.metricsEndpoint = endpoint }
}

func WithTracesEnabled(enabled bool) Option {
	return func(c *config) { c.tracesEnabled = enabled }
}

func WithMetricsEnabled(enabled bool) Option {
	return func(c *config) { c.metricsEnabled = enabled }
}

func WithMetricsReader(reader ...metricsdk.Reader) Option {
	return func(c *config) {
		c.metricsReaders = append(c.metricsReaders, reader...)
	}
}

func WithInsecure(insecure bool) Option {
	return func(c *config) { c.insecure = insecure }
}

func defaultConfig() *config {
	return &config{
		tracesEnabled:  false,
		metricsEnabled: false,
		insecure:       true,
	}
}

func (c *config) endpointFor(specific string) string {
	if specific != "" {
		return specific
	}
	return c.defaultEndpoint
}

func (c *config) applyEnvOverrides() {
	if v := os.Getenv("OTEL_SERVICE_NAME"); v != "" {
		c.serviceName = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		c.defaultEndpoint = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT"); v != "" {
		c.tracesEndpoint = v
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_METRICS_ENDPOINT"); v != "" {
		c.metricsEndpoint = v
	}

	c.autoEnable()

	if v := os.Getenv("OTEL_TRACES_EXPORTER"); v != "" {
		c.tracesEnabled = v != "none"
	}
	if v := os.Getenv("OTEL_METRICS_EXPORTER"); v != "" {
		c.metricsEnabled = v != "none"
	}
}

func (c *config) autoEnable() {
	if c.endpointFor(c.tracesEndpoint) != "" {
		c.tracesEnabled = true
	}
	if c.endpointFor(c.metricsEndpoint) != "" || len(c.metricsReaders) > 0 {
		c.metricsEnabled = true
	}
}
