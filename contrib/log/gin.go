package log

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"
	glog "github.com/lpphub/gost/log"
	"go.opentelemetry.io/otel/trace"
)

type requestLogConfig struct {
	skipPaths  map[string]struct{}
	fields     []func(c *gin.Context) map[string]any
	response   bool
	maxBodyLen int
}

type RequestLogOption func(*requestLogConfig)

func defaultConfig() *requestLogConfig {
	return &requestLogConfig{
		skipPaths: make(map[string]struct{}),
	}
}

func WithSkipPaths(paths ...string) RequestLogOption {
	return func(cfg *requestLogConfig) {
		for _, p := range paths {
			if p != "" {
				cfg.skipPaths[p] = struct{}{}
			}
		}
	}
}

func WithLogFields(fn func(c *gin.Context) map[string]any) RequestLogOption {
	return func(cfg *requestLogConfig) {
		cfg.fields = append(cfg.fields, fn)
	}
}

func WithResponseBody(maxLen int) RequestLogOption {
	return func(cfg *requestLogConfig) {
		cfg.response = true
		cfg.maxBodyLen = maxLen
	}
}

type bodyCaptureWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *bodyCaptureWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyCaptureWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func GinRequestLog(opts ...RequestLogOption) gin.HandlerFunc {
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

		ctx := c.Request.Context()

		span := trace.SpanFromContext(ctx)
		if sc := span.SpanContext(); sc.IsValid() {
			ctx = glog.WithTrace(ctx, sc.TraceID().String(), sc.SpanID().String())
		}

		c.Request = c.Request.WithContext(ctx)

		start := time.Now()

		var ww *bodyCaptureWriter
		if cfg.response {
			ww = &bodyCaptureWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
			c.Writer = ww
		}

		c.Next()

		ev := glog.Ctx(ctx).
			Info().
			Int("status", c.Writer.Status()).
			Int64("elapsed_ms", time.Since(start).Milliseconds()).
			Str("method", c.Request.Method).
			Str("path", c.Request.RequestURI)

		if ww != nil {
			body := ww.body.String()
			if cfg.maxBodyLen > 0 && len(body) > cfg.maxBodyLen {
				body = body[:cfg.maxBodyLen] + "...(truncated)"
			}
			ev = ev.Str("response", body)
		}

		for _, fn := range cfg.fields {
			for k, v := range fn(c) {
				ev = ev.Interface(k, v)
			}
		}

		ev.Msg("gin request")
	}
}
