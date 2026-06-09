package promexp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

var handler http.Handler

// NewReader creates a Prometheus metrics reader for use with otel.Init.
// The returned reader writes OTel metrics into an isolated Prometheus registry.
func NewReader() (metricsdk.Reader, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}

	handler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	return exp, nil
}

// Handler returns the HTTP handler that exposes the Prometheus metrics.
// Returns nil if NewReader has not been called.
func Handler() http.Handler {
	return handler
}
