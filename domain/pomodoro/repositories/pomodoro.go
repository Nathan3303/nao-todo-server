package repositories

import (
	"context"
	"time"

	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/valueobjects"
)

// Pomodoro 常用番茄工作仓库接口
type Pomodoro interface {
	// Create 创建常用番茄工作
	Create(
		ctx context.Context,
		vo *valueobjects.CreatePomodoro,
	) (*entities.Pomodoro, error)

	// Upsert 幂等写入：客户端指定 id 时创建/覆盖（LWW 判定 + create 冲突检测）
	// created=true 表示本次为新建
	Upsert(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreatePomodoro,
	) (*entities.Pomodoro, bool, error)

	// GetById 根据 ID 获取常用番茄工作
	GetById(ctx context.Context, userId int64, id int64) (*entities.Pomodoro, error)

	// Update 更新常用番茄工作（PATCH 语义）
	Update(
		ctx context.Context,
		userId int64,
		id int64,
		vo *valueobjects.UpdatePomodoro,
	) (*entities.Pomodoro, error)

	// Delete 删除常用番茄工作（软删除）
	Delete(ctx context.Context, userId int64, id int64) error

	// Archive 归档常用番茄工作
	Archive(ctx context.Context, userId int64, id int64) error

	// Unarchive 取消归档常用番茄工作
	Unarchive(ctx context.Context, userId int64, id int64) error

	// List 获取常用番茄工作列表
	List(
		ctx context.Context,
		userId int64,
		q *valueobjects.QueryPomodoro,
	) ([]*entities.Pomodoro, int64, error)

	// ListSync 增量同步列表：包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
	ListSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.Pomodoro, error)
}
