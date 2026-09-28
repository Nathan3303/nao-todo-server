package dbs

import (
	"context"

	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/derived"

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
	// 1. 嵌套保护：已在事务中则直接复用（派生写收集器同样复用，由最外层统一并入）
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx)
	}
	// 2. 开启事务并将事务句柄 + 事务级派生写收集器注入上下文
	txRecorder := derived.NewRecorder()
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(derived.WithRecorder(ctx, txRecorder), txKey{}, tx))
	})
	if err != nil {
		return err
	}
	// 3. 仅提交成功才并入请求级收集器（回滚丢弃，避免把未落库的派生写误报为可收敛版本）
	derived.Merge(ctx, txRecorder)
	return nil
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
