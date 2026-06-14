package otel

import (
	"context"

	gotel "github.com/lpphub/gost/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

func DBTelemetry(db *gorm.DB) *gorm.DB {
	tracer := gotel.Tracer("gorm")
	cb := db.Callback()

	cb.Create().Before("gorm:create").Register("otel:before_create", before(tracer, "CREATE"))
	cb.Create().After("gorm:create").Register("otel:after_create", after)

	cb.Query().Before("gorm:query").Register("otel:before_query", before(tracer, "SELECT"))
	cb.Query().After("gorm:query").Register("otel:after_query", after)

	cb.Update().Before("gorm:update").Register("otel:before_update", before(tracer, "UPDATE"))
	cb.Update().After("gorm:update").Register("otel:after_update", after)

	cb.Delete().Before("gorm:delete").Register("otel:before_delete", before(tracer, "DELETE"))
	cb.Delete().After("gorm:delete").Register("otel:after_delete", after)

	cb.Raw().Before("gorm:raw").Register("otel:before_raw", before(tracer, "RAW"))
	cb.Raw().After("gorm:raw").Register("otel:after_raw", after)

	return db
}

const otelSpanKey = "otel:span"

func before(tracer trace.Tracer, operation string) func(*gorm.DB) {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		if ctx == nil {
			ctx = context.Background()
		}

		tableName := ""
		if db.Statement.Schema != nil {
			tableName = db.Statement.Schema.Table
		}
		spanName := operation + " " + tableName

		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(
				attribute.String("db.operation", operation),
				attribute.String("db.sql.table", tableName),
			),
		)

		db.Statement.Context = ctx
		db.Set(otelSpanKey, span)
	}
}

func after(db *gorm.DB) {
	val, ok := db.Get(otelSpanKey)
	if !ok {
		return
	}
	span, ok := val.(trace.Span)
	if !ok {
		return
	}

	span.SetAttributes(
		attribute.Int64("db.rows_affected", db.RowsAffected),
	)

	if db.Error != nil {
		span.SetStatus(codes.Error, db.Error.Error())
		span.RecordError(db.Error)
	}

	span.End()
}
