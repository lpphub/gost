package otel

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
)

var (
	promHandler   http.Handler
	promHandlerMu sync.Mutex
)

func initPrometheus() (metricsdk.Reader, error) {
	reg := prometheus.NewRegistry()
	exp, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}

	promHandlerMu.Lock()
	promHandler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	promHandlerMu.Unlock()

	return exp, nil
}

func PrometheusHandler() http.Handler {
	promHandlerMu.Lock()
	defer promHandlerMu.Unlock()
	return promHandler
}
