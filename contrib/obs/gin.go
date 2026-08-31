package obs

import (
	"github.com/gin-gonic/gin"
	contriblog "github.com/lpphub/gost/contrib/log"
	contribotel "github.com/lpphub/gost/contrib/otel"
)

func GinObservability(serviceName string, logOpts ...contriblog.RequestLogOption) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		contribotel.GinTelemetry(serviceName),
		contriblog.GinRequestLog(logOpts...),
	}
}
