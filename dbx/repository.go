package dbx

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("dbx: record not found")

type BaseRepo[T any] struct {
	db *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{db: db}
}

func (r *BaseRepo[T]) DB() *gorm.DB {
	return r.db
}

func (r *BaseRepo[T]) First(ctx context.Context, id uint) (T, error) {
	var entity T
	if err := TxAwareDB(ctx, r.db).First(&entity, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return entity, ErrNotFound
		}
		return entity, fmt.Errorf("dbx: first by id %d: %w", id, err)
	}
	return entity, nil
}

func (r *BaseRepo[T]) FindByIDs(ctx context.Context, ids []uint) ([]T, error) {
	var entities []T
	if err := TxAwareDB(ctx, r.db).Where("id IN ?", ids).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("dbx: find by ids: %w", err)
	}
	return entities, nil
}

func (r *BaseRepo[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := TxAwareDB(ctx, r.db).Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("dbx: find all: %w", err)
	}
	return entities, nil
}

func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	if err := TxAwareDB(ctx, r.db).Create(entity).Error; err != nil {
		return fmt.Errorf("dbx: create: %w", err)
	}
	return nil
}

func (r *BaseRepo[T]) Update(ctx context.Context, id uint, updates map[string]any) error {
	if err := TxAwareDB(ctx, r.db).Model(new(T)).Where("id = ?", id).Updates(updates).Error; err != nil {
		return fmt.Errorf("dbx: update id %d: %w", id, err)
	}
	return nil
}

func (r *BaseRepo[T]) Delete(ctx context.Context, id uint) error {
	if err := TxAwareDB(ctx, r.db).Delete(new(T), id).Error; err != nil {
		return fmt.Errorf("dbx: delete id %d: %w", id, err)
	}
	return nil
}
