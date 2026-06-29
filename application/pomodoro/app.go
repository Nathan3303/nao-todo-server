package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/service"
	"naotodoserver/interfaces/types"
)

// PomodoroApp 专注应用应用层接口
type PomodoroApp interface {
	Create(ctx context.Context, req *types.CreatePomodoroReq) (*types.CreatePomodoroRes, error)
	Get(ctx context.Context, req *types.GetPomodoroReq) (*types.GetPomodoroRes, error)
	List(ctx context.Context, req *types.ListPomodoroReq) ([]*types.GetPomodoroRes, int64, error)
}

// PomodoroAppImpl 专注应用应用层实现
type PomodoroAppImpl struct {
	pomodoroDomain service.PomodoroDomain
}
