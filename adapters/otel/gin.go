package otel

import (
	"net/http"

	"github.com/gin-gonic/gin"
	gotel "github.com/lpphub/gost/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func GinTelemetry(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	defaults := []otelgin.Option{
		otelgin.WithTracerProvider(gotel.TracerProvider()),
		otelgin.WithMeterProvider(gotel.MeterProvider()),
	}
	return otelgin.Middleware(serviceName, append(defaults, opts...)...)
}

func RegisterMetricsEndpoint(r gin.IRouter, path string) {
	handler := gotel.PrometheusHandler()
	if handler == nil {
		return
	}
	r.GET(path, gin.WrapH(handler))
}

func PrometheusHandler() http.Handler {
	return gotel.PrometheusHandler()
}
