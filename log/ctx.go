package log

import (
	"context"

	"github.com/rs/zerolog"
)

var Ctx = zerolog.Ctx

func WithTrace(ctx context.Context, traceID, spanID string) context.Context {
	if traceID == "" {
		return ctx
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return zerolog.Ctx(ctx).With().
		Str("trace_id", traceID).
		Str("span_id", spanID).
		Logger().
		WithContext(ctx)
}
