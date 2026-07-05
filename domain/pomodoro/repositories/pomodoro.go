package repositories

import (
	"context"
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
}
