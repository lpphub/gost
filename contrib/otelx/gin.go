package otelx

import (
	"github.com/gin-gonic/gin"
	"github.com/lpphub/gost/otel"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func GinMiddleware(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	defaults := []otelgin.Option{
		otelgin.WithTracerProvider(otel.TracerProvider()),
		otelgin.WithMeterProvider(otel.MeterProvider()),
	}
	return otelgin.Middleware(serviceName, append(defaults, opts...)...)
}
