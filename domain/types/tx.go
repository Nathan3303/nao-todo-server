package types

import "context"

// TxManager 事务管理端口
// 用于将多个仓储操作包裹在单一事务中
type TxManager interface {
	// Do 在事务中执行 fn
	// fn 返回错误时事务回滚，返回 nil 时提交
	// @param ctx 上下文
	// @param fn 事务内执行的函数，其 ctx 携带事务句柄
	// @return error 错误
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
