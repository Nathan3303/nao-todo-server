package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/service"
	"naotodoserver/interfaces/types"
)

// PomodoroApp 专注应用应用层接口
type PomodoroApp interface {
	// Create 创建番茄工作记录
	Create(
		ctx context.Context,
		req *types.CreatePomodoroRecordReq,
	) (*types.CreatePomodoroRecordRes, error)

	// Get 获取番茄工作记录
	Get(
		ctx context.Context,
		req *types.GetPomodoroRecordReq,
	) (*types.GetPomodoroRecordRes, error)

	// List 获取番茄工作记录列表
	List(
		ctx context.Context,
		req *types.ListPomodoroRecordReq,
	) ([]*types.GetPomodoroRecordRes, int64, error)
}

// PomodoroAppImpl 专注应用应用层实现
type PomodoroAppImpl struct {
	pomodoroDomain service.PomodoroDomain
}
