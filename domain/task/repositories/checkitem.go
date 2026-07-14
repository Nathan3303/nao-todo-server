package repositories

import (
	"context"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// TaskCheckItem 任务检查项仓库接口
type TaskCheckItem interface {
	// GetCheckItemById 获取单个任务检查项信息
	GetCheckItemById(
		ctx context.Context,
		userId int64,
		checkItemId int64,
	) (*entities.TaskCheckItem, error)

	// CreateCheckItem 创建任务检查项
	CreateCheckItem(
		ctx context.Context,
		userId int64,
		vo *valueobjects.CreateTaskCheckItem,
	) (*entities.TaskCheckItem, error)

	// UpdateCheckItem 更新任务检查项
	UpdateCheckItem(
		ctx context.Context,
		userId int64,
		checkItemId int64,
		vo *valueobjects.UpdateTaskCheckItem,
	) error

	// DeleteCheckItem 删除任务检查项
	DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error

	// ListCheckItems 获取任务检查项列表
	ListCheckItems(ctx context.Context, userId, taskId int64) ([]*entities.TaskCheckItem, error)

	// GetMaxCheckItemSortId 获取任务检查项最大排序 ID
	GetMaxCheckItemSortId(ctx context.Context, userId, taskId int64) uint16

	// BatchUpdateCheckItems 批量更新任务检查项
	BatchUpdateCheckItems(
		ctx context.Context,
		userId int64,
		vos []*valueobjects.BatchUpdateTaskCheckItem,
	) ([]*entities.TaskCheckItem, error)
}
