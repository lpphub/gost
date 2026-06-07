package dbx

import (
	"context"

	"gorm.io/gorm"
)

type contextTxKey struct{}

// WithTx 将事务 DB 注入 context
func WithTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, contextTxKey{}, tx)
}

// TxFromContext 从 context 中获取事务 DB
func TxFromContext(ctx context.Context) *gorm.DB {
	tx, _ := ctx.Value(contextTxKey{}).(*gorm.DB)
	return tx
}

// TxAwareDB 优先返回事务DB, 若不存在事务则返回指定db
func TxAwareDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx := TxFromContext(ctx); tx != nil {
		return tx
	}
	return db.WithContext(ctx)
}

// InTransaction 在事务中执行 fn, db 为 nil 时直接执行 fn（无事务）
func InTransaction(ctx context.Context, db *gorm.DB, fn func(context.Context) error) error {
	if db == nil {
		return fn(ctx)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		txCtx := WithTx(ctx, tx)
		return fn(txCtx)
	})
}
