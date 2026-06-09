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

// RegisterMetricsEndpoint registers a Prometheus metrics handler at the given path.
// Pass promexp.Handler() as the handler if you're using the Prometheus pull exporter.
func RegisterMetricsEndpoint(r gin.IRouter, path string, handler http.Handler) {
	if handler == nil {
		return
	}
	r.GET(path, gin.WrapH(handler))
}
