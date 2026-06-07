package dbx

import (
	"context"

	"gorm.io/gorm"
)

// BaseRepo 通用 Repository 基础结构
type BaseRepo[T any] struct {
	db *gorm.DB
}

func NewBaseRepo[T any](db *gorm.DB) *BaseRepo[T] {
	return &BaseRepo[T]{db: db}
}

// DB 获取数据库实例
func (r *BaseRepo[T]) DB() *gorm.DB {
	return r.db
}

// First 根据 ID 获取单条记录
func (r *BaseRepo[T]) First(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := r.db.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

// FindByIDs 批量获取记录
func (r *BaseRepo[T]) FindByIDs(ctx context.Context, ids []uint) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// FindAll 获取所有记录
func (r *BaseRepo[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	if err := r.db.WithContext(ctx).Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

// Create 创建记录（ctx 事务感知）
func (r *BaseRepo[T]) Create(ctx context.Context, entity *T) error {
	return TxAwareDB(ctx, r.db).Create(entity).Error
}

// Update 更新记录（ctx 事务感知）
func (r *BaseRepo[T]) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return TxAwareDB(ctx, r.db).Model(new(T)).Where("id = ?", id).Updates(updates).Error
}

// Delete 删除记录（ctx 事务感知）
func (r *BaseRepo[T]) Delete(ctx context.Context, id uint) error {
	return TxAwareDB(ctx, r.db).Delete(new(T), id).Error
}
