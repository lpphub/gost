package otel

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
)

func RedisTelemetry(client *redis.Client) *redis.Client {
	if err := redisotel.InstrumentTracing(client); err != nil {
		otel.Handle(err)
	}
	return client
}
