package dbx

import (
	"context"

	"gorm.io/gorm"
)

type ctxTxKey struct{}

func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, ctxTxKey{}, tx)
}

func TxFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(ctxTxKey{}).(*gorm.DB)
	return tx
}

func TxAwareDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}

func InTransaction(ctx context.Context, db *gorm.DB, fn func(context.Context) error) error {
	if db == nil {
		return fn(ctx)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		txCtx := WithTx(ctx, tx)
		return fn(txCtx)
	})
}
