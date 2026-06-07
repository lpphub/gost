package logging

import (
	"context"
	"testing"
)

func TestCtxInfo(t *testing.T) {
	Init()

	t.Run("CtxLog", func(t *testing.T) {
		ctx := context.Background()
		Info(ctx, "test111")

		ctx = WithFields(ctx, Str("requestId", "ABC123"))

		Info(ctx, "test222")

		L().Error(ctx).Str("field", "0000").Msg("test333")

		ctx = WithFields(ctx, Str("field", "1111"))

		Info(ctx, "test444")

		L().Info(ctx).Str("field", "2222").Msg("test555")
	})
}
