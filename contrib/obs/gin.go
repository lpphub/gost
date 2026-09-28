package obs

import (
	"github.com/gin-gonic/gin"
	"github.com/lpphub/gost/contrib/log"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// GinObs 返回 Gin 的中间件组合：先追踪中间件、再请求日志，顺序已编排好。
func GinObs(serviceName string, logOpts ...log.RequestLogOption) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		otelgin.Middleware(serviceName),
		log.GinRequestLog(logOpts...),
	}
}

func Gin(serviceName string) gin.HandlerFunc {
	return otelgin.Middleware(serviceName)
}
