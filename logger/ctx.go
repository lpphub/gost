package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type loggerKey struct{}

func ensureCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func WithTrace(ctx context.Context, traceID, spanID string) context.Context {
	ctx = ensureCtx(ctx)
	if traceID == "" {
		return ctx
	}
	l := *Ctx(ctx)
	l = l.With().Str("trace_id", traceID).Str("span_id", spanID).Logger()
	return context.WithValue(ctx, loggerKey{}, &l)
}

func WithCtx(ctx context.Context, fields ...Field) context.Context {
	ctx = ensureCtx(ctx)
	if len(fields) == 0 {
		return ctx
	}
	l := *Ctx(ctx)
	sub := l.With()
	for _, f := range fields {
		sub = f.ApplyCtx(sub)
	}
	result := sub.Logger()
	return context.WithValue(ctx, loggerKey{}, &result)
}

func Ctx(ctx context.Context) *zerolog.Logger {
	if ctx == nil {
		return L()
	}
	if l, ok := ctx.Value(loggerKey{}).(*zerolog.Logger); ok {
		return l
	}
	return L()
}
