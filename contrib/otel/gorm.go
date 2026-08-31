package otel

import (
	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm"
)

func DBTelemetry(db *gorm.DB) *gorm.DB {
	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		otel.Handle(err)
	}
	return db
}
