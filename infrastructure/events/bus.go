package events

import (
	"context"
	"sync"

	"naotodoserver/domain/types"
)

// InMemoryBus 进程内同步事件总线（计数事件）
// 订阅者注册表 + 同步 Dispatch：事件在写事务内、提交前同步分发；
// 任一订阅者返回错误 ⇒ 发布返回错误 ⇒ 外层事务回滚（计数与主写永不分离，ADR §4.1/§4.2）。
type InMemoryBus struct {
	mu       sync.RWMutex
	handlers []func(ctx context.Context, event types.CountEvent) error
}

// 编译期接口实现断言
var _ types.CountEventPublisher = (*InMemoryBus)(nil)

// NewInMemoryBus 创建内存事件总线
func NewInMemoryBus() *InMemoryBus {
	return &InMemoryBus{}
}

// Subscribe 注册订阅者（按注册顺序同步分发；可注册多个）
func (b *InMemoryBus) Subscribe(h func(ctx context.Context, event types.CountEvent) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers = append(b.handlers, h)
}

// PublishCountEvent 同步分发计数事件
// 快照订阅者列表后依次调用；任一订阅者失败立即返回错误（触发调用方事务回滚）。
func (b *InMemoryBus) PublishCountEvent(ctx context.Context, event types.CountEvent) error {
	b.mu.RLock()
	handlers := make([]func(ctx context.Context, event types.CountEvent) error, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()
	for _, h := range handlers {
		if err := h(ctx, event); err != nil {
			return err
		}
	}
	return nil
}
