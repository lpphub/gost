package obs

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

// DB 给 GORM 挂上追踪插件，失败只上报给 otel 错误处理器，不返回。
func DB(db *gorm.DB, opts ...otelgorm.Option) *gorm.DB {
	if err := db.Use(otelgorm.NewPlugin(opts...)); err != nil {
		otel.Handle(err)
	}
	return db
}
