package log

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
	if traceID == "" {
		return ctx
	}
	return WithCtx(ctx, CtxStr("trace_id", traceID), CtxStr("span_id", spanID))
}

func WithCtx(ctx context.Context, fields ...CtxField) context.Context {
	ctx = ensureCtx(ctx)
	if len(fields) == 0 {
		return ctx
	}
	l := *Ctx(ctx)
	sub := l.With()
	for _, f := range fields {
		sub = f(sub)
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
