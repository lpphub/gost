package logx

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/gost/logger"
	"go.opentelemetry.io/otel/trace"
)

type traceLogConfig struct {
	skipPaths map[string]struct{}
}

type TraceLogOption func(*traceLogConfig)

func defaultTraceLogConfig() *traceLogConfig {
	return &traceLogConfig{
		skipPaths: make(map[string]struct{}),
	}
}

func WithSkipPaths(paths ...string) TraceLogOption {
	return func(cfg *traceLogConfig) {
		for _, p := range paths {
			if p != "" {
				cfg.skipPaths[p] = struct{}{}
			}
		}
	}
}

func GinTraceLog(opts ...TraceLogOption) gin.HandlerFunc {
	cfg := defaultTraceLogConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		path := c.FullPath()
		if _, ok := cfg.skipPaths[path]; ok {
			c.Next()
			return
		}

		ctx := c.Request.Context()

		span := trace.SpanFromContext(ctx)
		if sc := span.SpanContext(); sc.IsValid() {
			ctx = logger.WithTrace(ctx, sc.TraceID().String(), sc.SpanID().String())
		}

		c.Request = c.Request.WithContext(ctx)

		start := time.Now()
		c.Next()

		logger.Ctx(ctx).
			Info().
			Int("status", c.Writer.Status()).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Str("method", c.Request.Method).
			Str("path", c.Request.RequestURI).
			Msg("gin request")
	}
}
