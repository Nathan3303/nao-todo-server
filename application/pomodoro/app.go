package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/service"
	"naotodoserver/interfaces/types"
)

// PomodoroApp 专注应用应用层接口
type PomodoroApp interface {
	// --- PomodoroRecord ---

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

	// --- Pomodoro ---

	// CreatePomodoro 创建常用番茄工作
	CreatePomodoro(
		ctx context.Context,
		req *types.CreatePomodoroReq,
	) (*types.CreatePomodoroRes, error)

	// GetPomodoro 获取常用番茄工作
	GetPomodoro(
		ctx context.Context,
		req *types.GetPomodoroReq,
	) (*types.PomodoroRes, error)

	// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
	UpdatePomodoro(
		ctx context.Context,
		id string,
		req *types.UpdatePomodoroReq,
	) error

	// DeletePomodoro 删除常用番茄工作（软删除）
	DeletePomodoro(ctx context.Context, id string) error

	// ArchivePomodoro 归档常用番茄工作
	ArchivePomodoro(ctx context.Context, id string) error

	// UnarchivePomodoro 取消归档常用番茄工作
	UnarchivePomodoro(ctx context.Context, id string) error

	// ListPomodoro 获取常用番茄工作列表
	ListPomodoro(
		ctx context.Context,
		req *types.ListPomodoroReq,
	) (types.ListPomodoroRes, int64, error)
}

// PomodoroAppImpl 专注应用应用层实现
type PomodoroAppImpl struct {
	pomodoroDomain      service.PomodoroDomain
	pomodoroRecordRepo  repositories.PomodoroRecord
	pomodoroRepo        repositories.Pomodoro
}
