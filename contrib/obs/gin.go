package obs

import (
	"github.com/gin-gonic/gin"
	contriblog "github.com/lpphub/gost/contrib/log"
	contribotel "github.com/lpphub/gost/contrib/otel"
)

// GinObservability 按顺序返回 trace + request log 两个观测中间件。
func GinObservability(serviceName string, logOpts ...contriblog.RequestLogOption) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		contribotel.GinTraceMiddleware(serviceName),
		contriblog.GinRequestLog(logOpts...),
	}
}
