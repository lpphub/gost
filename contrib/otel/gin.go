package otel

import (
	"github.com/gin-gonic/gin"
	gotel "github.com/lpphub/gost/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GinTraceMiddleware returns a tracing middleware for Gin.
func GinTraceMiddleware(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	defaults := []otelgin.Option{
		otelgin.WithTracerProvider(gotel.TracerProvider()),
		otelgin.WithMeterProvider(gotel.MeterProvider()),
	}
	return otelgin.Middleware(serviceName, append(defaults, opts...)...)
}
