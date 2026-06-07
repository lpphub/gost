package logx

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lpphub/goweb/pkg/logging"
	"github.com/oklog/ulid/v2"
)

const (
	// HeaderRequestID 请求中携带的 requestID Header
	HeaderRequestID = "X-Request-ID"
	ctxKeyRequestID = "requestId"
)

type accessLogConfig struct {
	skipPaths map[string]struct{}
}

type AccessLogOption func(*accessLogConfig)

func defaultConfig() *accessLogConfig {
	return &accessLogConfig{
		skipPaths: make(map[string]struct{}),
	}
}

// WithSkipPaths 跳过指定 Gin 路由，如 /health
func WithSkipPaths(paths ...string) AccessLogOption {
	return func(cfg *accessLogConfig) {
		for _, p := range paths {
			if p != "" {
				cfg.skipPaths[p] = struct{}{}
			}
		}
	}
}

// GinAccessLog Gin 请求访问日志中间件（支持跳过路径）
func GinAccessLog(opts ...AccessLogOption) gin.HandlerFunc {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	return func(c *gin.Context) {
		path := c.FullPath()
		if _, ok := cfg.skipPaths[path]; ok {
			c.Next()
			return
		}

		start := time.Now()

		// 解析或生成 requestId
		requestID := resolveRequestID(c)

		// 注入 gin context
		c.Set(ctxKeyRequestID, requestID)

		// 注入 context 中
		ctx := logging.WithFields(c.Request.Context(), logging.Str(ctxKeyRequestID, requestID))

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		logging.L().Info(ctx).
			Int("status", c.Writer.Status()).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Str("method", c.Request.Method).
			Str("path", c.Request.RequestURI).
			Msg("gin access")
	}
}

func resolveRequestID(c *gin.Context) string {
	if c.Request != nil {
		if logID := c.GetHeader(HeaderRequestID); logID != "" {
			return logID
		}
	}
	return GenerateRequestID()
}

func GenerateRequestID() string {
	return ulid.Make().String()
}

// GetRequestID 从 gin.Context 获取 requestId
func GetRequestID(c *gin.Context) string {
	if v, ok := c.Get(ctxKeyRequestID); ok {
		if s, ok2 := v.(string); ok2 {
			return s
		}
	}
	return ""
}
