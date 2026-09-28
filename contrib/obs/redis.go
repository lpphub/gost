package obs

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

// Redis 给 go-redis 客户端挂上追踪埋点，失败只上报给 otel 错误处理器。
func Redis(client *redis.Client, opts ...redisotel.TracingOption) *redis.Client {
	if err := redisotel.InstrumentTracing(client, opts...); err != nil {
		otel.Handle(err)
	}
	return client
}
