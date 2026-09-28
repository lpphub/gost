package otel

import "os"

type settings struct {
	service           string
	protocol          string
	tracesExporter    string
	metricsExporter   string
	hasTraceEndpoint  bool
	hasMetricEndpoint bool
}

func settingsFrom(service string) settings {
	return settings{
		service:           service,
		protocol:          protocolFromEnv(os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")),
		tracesExporter:    os.Getenv("OTEL_TRACES_EXPORTER"),
		metricsExporter:   os.Getenv("OTEL_METRICS_EXPORTER"),
		hasTraceEndpoint:  hasEndpoint("TRACES"),
		hasMetricEndpoint: hasEndpoint("METRICS"),
	}
}

func protocolFromEnv(v string) string {
	switch v {
	case "", "http", "http/protobuf":
		return ProtocolHTTP
	}
	return v
}

func hasEndpoint(signal string) bool {
	return os.Getenv("OTEL_EXPORTER_OTLP_"+signal+"_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != ""
}
