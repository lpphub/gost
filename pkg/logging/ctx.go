package logging

import (
	"context"
)

type ctxFieldsKey struct{}

// WithFields 绑定 fields 到 ctx
func WithFields(ctx context.Context, fields ...Field) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	if len(fields) == 0 {
		return ctx
	}

	existing, _ := ctx.Value(ctxFieldsKey{}).([]Field)
	all := append(existing, fields...)

	return context.WithValue(ctx, ctxFieldsKey{}, all)
}

// FieldsFrom 获取 ctx 中绑定的 fields
func FieldsFrom(ctx context.Context) []Field {
	if ctx == nil {
		return nil
	}

	fields, _ := ctx.Value(ctxFieldsKey{}).([]Field)
	return fields
}
