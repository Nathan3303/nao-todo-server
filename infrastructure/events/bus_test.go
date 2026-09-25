package events

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"naotodoserver/domain/types"
)

func TestInMemoryBus_DispatchInOrderAndErrorPropagation(t *testing.T) {
	bus := NewInMemoryBus()
	var order []types.CountEventType
	markErr := errors.New("订阅者失败")

	// 订阅者 1：记录事件
	bus.Subscribe(func(ctx context.Context, event types.CountEvent) error {
		order = append(order, event.Type)
		return nil
	})
	// 订阅者 2：记录事件后返回错误（模拟计数更新失败 ⇒ 外层事务应回滚）
	bus.Subscribe(func(ctx context.Context, event types.CountEvent) error {
		order = append(order, event.Type)
		return markErr
	})

	event := types.CountEvent{
		Type:      types.CountEventTaskCountChanged,
		UserId:    1,
		ProjectId: 100,
		Delta:     1,
	}
	err := bus.PublishCountEvent(context.Background(), event)
	if !errors.Is(err, markErr) {
		t.Fatalf("PublishCountEvent 应返回订阅者错误，got %v", err)
	}
	// 两个订阅者均被调用（同步分发，先注册先执行）
	want := []types.CountEventType{
		types.CountEventTaskCountChanged,
		types.CountEventTaskCountChanged,
	}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("分发顺序错误: got %v, want %v", order, want)
	}
}

func TestInMemoryBus_NoSubscriberIsNoop(t *testing.T) {
	bus := NewInMemoryBus()
	if err := bus.PublishCountEvent(context.Background(), types.CountEvent{
		Type: types.CountEventCommentChanged,
	}); err != nil {
		t.Fatalf("无订阅者应返回 nil，got %v", err)
	}
}

// 编译期断言：总线实现 CountEventPublisher 端口
var _ types.CountEventPublisher = (*InMemoryBus)(nil)
