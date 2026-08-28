package otel

import (
	"context"
	"errors"
	"fmt"

	gotel "github.com/lpphub/gost/otel"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

func RedisTelemetry(client *redis.Client) *redis.Client {
	client.AddHook(&otelRedisHook{
		client: client,
		tracer: gotel.Tracer("redis"),
	})
	return client
}

type otelRedisHook struct {
	client *redis.Client
	tracer trace.Tracer
}

func (h *otelRedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *otelRedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		ctx, span := h.tracer.Start(ctx, "redis."+cmd.Name(),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(append(
				h.baseAttrs(),
				attribute.String("db.operation", cmd.Name()),
			)...),
		)
		defer span.End()

		err := next(ctx, cmd)
		h.recordError(span, err)
		return err
	}
}

func (h *otelRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		ctx, span := h.tracer.Start(ctx, fmt.Sprintf("redis.pipeline(%d)", len(cmds)),
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(append(
				h.baseAttrs(),
				attribute.String("db.operation", "pipeline"),
				attribute.Int("db.redis.pipeline_size", len(cmds)),
			)...),
		)
		defer span.End()

		err := next(ctx, cmds)
		h.recordError(span, err)
		return err
	}
}

func (h *otelRedisHook) baseAttrs() []attribute.KeyValue {
	attrs := make([]attribute.KeyValue, 3, 5)
	attrs[0] = semconv.DBSystemRedis
	attrs[1] = semconv.ServerAddressKey.String(h.client.Options().Addr)
	attrs[2] = attribute.Int("db.redis.database_index", h.client.Options().DB)
	return attrs
}

func (h *otelRedisHook) recordError(span trace.Span, err error) {
	if err != nil && !errors.Is(err, redis.Nil) {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}
}
