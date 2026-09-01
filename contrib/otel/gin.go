package otel

import (
	"github.com/gin-gonic/gin"
	"github.com/lpphub/gost/contrib/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func GinTelemetry(serviceName string, opts ...otelgin.Option) gin.HandlerFunc {
	return otelgin.Middleware(serviceName, opts...)
}

func GinObservability(serviceName string, logOpts ...log.RequestLogOption) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		GinTelemetry(serviceName),
		log.GinRequestLog(logOpts...),
	}
}
