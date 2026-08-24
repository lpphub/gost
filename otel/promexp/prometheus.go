package promexp

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

var handler http.Handler

func NewReader() (metricsdk.Reader, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}

	handler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})

	return exp, nil
}

func Handler() http.Handler {
	return handler
}
