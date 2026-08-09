package pomodoro

import (
	"context"

	"naotodoserver/application/pomodoro/dto"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/service"
)

// PomodoroApp 专注应用应用层接口
type PomodoroApp interface {
	// --- PomodoroRecord ---

	// Create 创建番茄工作记录
	Create(
		ctx context.Context,
		userId int64,
		req *dto.CreatePomodoroRecordReq,
	) (*dto.CreatePomodoroRecordRes, error)

	// Get 获取番茄工作记录
	Get(
		ctx context.Context,
		userId int64,
		req *dto.GetPomodoroRecordReq,
	) (*dto.GetPomodoroRecordRes, error)

	// List 获取番茄工作记录列表
	List(
		ctx context.Context,
		userId int64,
		req *dto.ListPomodoroRecordReq,
	) ([]*dto.GetPomodoroRecordRes, int64, error)

	// ListSync 增量同步番茄工作记录列表（包含软删墓碑，updated_at 游标稳定排序分页）
	ListSync(
		ctx context.Context,
		userId int64,
		req *dto.ListPomodoroRecordReq,
	) ([]*dto.GetPomodoroRecordRes, error)

	// --- Pomodoro ---

	// CreatePomodoro 创建常用番茄工作
	CreatePomodoro(
		ctx context.Context,
		userId int64,
		req *dto.CreatePomodoroReq,
	) (*dto.CreatePomodoroRes, error)

	// GetPomodoro 获取常用番茄工作
	GetPomodoro(
		ctx context.Context,
		userId int64,
		req *dto.GetPomodoroReq,
	) (*dto.PomodoroRes, error)

	// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
	UpdatePomodoro(
		ctx context.Context,
		userId int64,
		id string,
		req *dto.UpdatePomodoroReq,
	) error

	// DeletePomodoro 删除常用番茄工作（软删除）
	DeletePomodoro(ctx context.Context, userId int64, id string) error

	// ArchivePomodoro 归档常用番茄工作
	ArchivePomodoro(ctx context.Context, userId int64, id string) error

	// UnarchivePomodoro 取消归档常用番茄工作
	UnarchivePomodoro(ctx context.Context, userId int64, id string) error

	// ListPomodoro 获取常用番茄工作列表
	ListPomodoro(
		ctx context.Context,
		userId int64,
		req *dto.ListPomodoroReq,
	) (dto.ListPomodoroRes, int64, error)

	// ListPomodoroSync 增量同步常用番茄工作列表（包含软删墓碑，updated_at 游标稳定排序分页）
	ListPomodoroSync(
		ctx context.Context,
		userId int64,
		req *dto.ListPomodoroReq,
	) (dto.ListPomodoroRes, error)
}

// PomodoroAppImpl 专注应用应用层实现
type PomodoroAppImpl struct {
	pomodoroDomain     service.PomodoroDomain
	pomodoroRecordRepo repositories.PomodoroRecord
	pomodoroRepo       repositories.Pomodoro
}
