package dbx

import (
	"context"

	"gorm.io/gorm"
)

type BaseRepo[T any] struct {
	db *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{db: db}
}

func (r *BaseRepo[T]) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepo[T]) First(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *BaseRepo[T]) FindByIDs(ctx context.Context, ids []uint) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepo[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	return TxAwareDB(ctx, r.db).Create(entity).Error
}

func (r *BaseRepo[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return TxAwareDB(ctx, r.db).Model(new(T)).Where("id = ?", id).Updates(updates).Error
}

func (r *BaseRepo[T]) Delete(ctx context.Context, id uint) error {
	return TxAwareDB(ctx, r.db).Delete(new(T), id).Error
}
