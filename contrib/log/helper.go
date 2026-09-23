package log

import (
	"context"

	"github.com/lpphub/gost/log"
	"github.com/lpphub/gost/otel"
	"github.com/rs/zerolog"
)

func truncate(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen] + " ...[truncated]"
	}
	return s
}

func withCaller(ev *zerolog.Event, skip int) *zerolog.Event {
	if skip > 0 {
		return ev.Caller(skip)
	}
	return ev
}

func WithSpan(ctx context.Context, name string) context.Context {
	_ctx, span := otel.Span(ctx, name)
	defer span.End()

	sc := span.SpanContext()
	return log.WithTrace(_ctx, sc.TraceID().String(), sc.SpanID().String())
}
