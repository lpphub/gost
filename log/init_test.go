package log

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
)

func TestCtxInfo(t *testing.T) {
	Init(WithLevel(zerolog.DebugLevel))

	t.Run("no fields returns global logger", func(t *testing.T) {
		ctx := context.Background()
		l := Ctx(ctx)
		if l == nil {
			t.Fatal("expected non-nil logger")
		}
		Info(ctx, "no fields - global logger")
	})

	t.Run("WithCtx fields", func(t *testing.T) {
		ctx := WithCtx(context.Background(), CtxStr("requestId", "ABC123"))
		Info(ctx, "with requestId")

		ctx = WithCtx(ctx, CtxStr("uid", "42"))
		Info(ctx, "with requestId and uid")
	})

	t.Run("WithTrace", func(t *testing.T) {
		ctx := WithTrace(context.Background(), "trace123", "span456")
		Info(ctx, "with trace")
	})

	t.Run("WithTrace + WithCtx", func(t *testing.T) {
		ctx := WithTrace(context.Background(), "trace123", "span456")
		ctx = WithCtx(ctx, CtxStr("uid", "42"))
		Info(ctx, "with trace and field")
	})

	t.Run("native chain", func(t *testing.T) {
		ctx := WithCtx(context.Background(), CtxStr("key", "val"))
		Ctx(ctx).Info().Str("extra", "data").Msg("native chain")
	})
}
