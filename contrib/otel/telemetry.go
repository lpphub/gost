package otel

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

// DBTelemetry 给 GORM 挂上追踪插件，失败只上报给 otel 错误处理器，不返回。
func DBTelemetry(db *gorm.DB, opts ...otelgorm.Option) *gorm.DB {
	if err := db.Use(otelgorm.NewPlugin(opts...)); err != nil {
		otel.Handle(err)
	}
	return db
}

// RedisTelemetry 给 go-redis 客户端挂上追踪埋点，失败只上报给 otel 错误处理器。
func RedisTelemetry(client *redis.Client, opts ...redisotel.TracingOption) *redis.Client {
	if err := redisotel.InstrumentTracing(client, opts...); err != nil {
		otel.Handle(err)
	}
	return client
}
