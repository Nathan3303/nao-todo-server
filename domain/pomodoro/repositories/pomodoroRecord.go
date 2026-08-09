package repositories

import (
	"context"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// PomodoroRecord Pomodoro 任务记录接口
type PomodoroRecord interface {
	// Create 创建番茄工作记录
	Create(
		ctx context.Context,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, error)

	// Upsert 幂等写入：客户端指定 id 时创建/覆盖（LWW 判定 + create 冲突检测）
	// created=true 表示本次为新建
	Upsert(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoroRecord,
	) (*entities.PomodoroRecord, bool, error)

	// GetById 根据 ID 获取番茄工作记录
	GetById(ctx context.Context, userId int64, id int64) (*entities.PomodoroRecord, error)

	// List 获取番茄工作记录列表
	List(
		ctx context.Context,
		userId int64,
		q *valueobjects.QueryPomodoroRecord,
	) ([]*entities.PomodoroRecord, int64, error)

	// ListSync 增量同步列表：包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
	ListSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.PomodoroRecord, error)
}
