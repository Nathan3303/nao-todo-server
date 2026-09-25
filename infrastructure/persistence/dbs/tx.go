package dbs

import (
	"context"
	"naotodoserver/domain/types"

	"gorm.io/gorm"
)

// txKey 事务句柄在上下文中的键类型（私有，避免键冲突）
type txKey struct{}

type txManagerImpl struct {
	db *gorm.DB
}

// NewTxManager 创建事务管理器实现
// @param db 数据库连接
// @return types.TxManager 事务管理端口
func NewTxManager(db *gorm.DB) types.TxManager {
	return &txManagerImpl{db: db}
}

// Do 在事务中执行 fn
// 若上下文中已存在事务句柄，则复用外层事务，不再开启新事务
// @param ctx 上下文
// @param fn 事务内执行的函数，其 ctx 携带事务句柄
// @return error 错误
func (m *txManagerImpl) Do(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	// 1. 嵌套保护：已在事务中则直接复用
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	// 2. 开启事务并将事务句柄注入上下文
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// DBFrom 从上下文中取出事务句柄
// 取不到时返回 fallback，使仓储方法在事务内外均可工作
// @param ctx 上下文
// @param fallback 无事务时使用的数据库连接
// @return *gorm.DB 数据库连接
func DBFrom(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return fallback
}
