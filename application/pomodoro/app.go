package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/service"
	"naotodoserver/interfaces/types"
)

type PomodoroApp interface {
	CreatePomodoro(ctx context.Context, req *types.CreatePomodoroReq) (*types.CreatePomodoroRes, error)
	GetPomodoro(ctx context.Context, req *types.GetPomodoroReq) (*types.GetPomodoroRes, error)
	ListPomodoro(ctx context.Context, req *types.ListPomodoroReq) ([]*types.GetPomodoroRes, int64, error)
}

type PomodoroAppImpl struct {
	pomodoroDomain service.PomodoroDomain
}
