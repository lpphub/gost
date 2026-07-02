package log

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestCtxInfo(t *testing.T) {
	var buf bytes.Buffer
	Init(WithLevel(zerolog.DebugLevel), WithWriter(&buf))

	t.Run("no fields returns global logger", func(t *testing.T) {
		buf.Reset()
		Ctx(context.Background()).Info().Msg("no fields")
		if !strings.Contains(buf.String(), "no fields") {
			t.Error("expected log output")
		}
	})

	t.Run("WithTrace", func(t *testing.T) {
		buf.Reset()
		ctx := WithTrace(context.Background(), "trace123", "span456")
		Ctx(ctx).Info().Msg("with trace")
		out := buf.String()
		if !strings.Contains(out, `"trace_id":"trace123"`) {
			t.Error("expected trace_id in output")
		}
		if !strings.Contains(out, `"span_id":"span456"`) {
			t.Error("expected span_id in output")
		}
	})

	t.Run("Caller", func(t *testing.T) {
		buf.Reset()
		Ctx(context.Background()).Info().Msg("with caller")
		out := buf.String()
		if !strings.Contains(out, `"caller":"log/init_test.go:`) {
			t.Error("expected caller file:line in output, got:", out)
		}
	})
}
