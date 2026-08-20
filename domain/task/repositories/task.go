package repositories

import (
	"context"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
)

// Task 任务仓库接口
type Task interface {
	// GetById 获取单个任务信息
	// includeDeleted 为 true 时可查询到已软删除的任务
	GetById(
		ctx context.Context,
		userId int64,
		taskId int64,
		includeDeleted bool,
	) (*entities.Task, error)

	// Create 创建任务
	Create(
		ctx context.Context,
		userId int64,
		createTaskValueObject *valueobjects.CreateTask,
	) (*entities.Task, error)

	// Upsert 幂等写入：客户端指定 id 时创建/覆盖（LWW 判定 + create 冲突检测）
	// created=true 表示本次为新建（调用方需初始化偏好等附属记录）
	Upsert(
		ctx context.Context,
		userId int64,
		createTaskValueObject *valueobjects.CreateTask,
	) (*entities.Task, bool, error)

	// Update 更新任务
	Update(
		ctx context.Context,
		userId int64,
		taskId int64,
		updateTaskValueObject *valueobjects.UpdateTask,
	) error

	// Delete 删除任务
	Delete(ctx context.Context, userId int64, taskId int64) error

	// Restore 恢复任务
	Restore(ctx context.Context, userId int64, taskId int64) error

	// List 获取任务列表
	List(
		ctx context.Context,
		userId int64,
		query *valueobjects.QueryTask,
		pagination *valueobjects.Pagination,
	) ([]*entities.Task, *valueobjects.Pagination, error)

	// ListSync 增量同步列表：包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页
	ListSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.Task, error)

	// Snooze 稍后提醒
	Snooze(ctx context.Context, userId int64, taskId int64, remindAt string) error

	// GetMaxSortId 获取任务最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64) uint16

	// GetDueReminders 获取到期提醒任务
	GetDueReminders(ctx context.Context) ([]*entities.Task, error)

	// ClearRemindRepeat 清除提醒重复规则
	// 仅当任务 remind_at 仍为 expectedRemindAt 时生效（CAS 防与 Snooze 竞态）；
	// 返回是否实际变更（false 表示提醒已被其他路径改期/删除）
	ClearRemindRepeat(ctx context.Context, taskId int64, expectedRemindAt time.Time) (bool, error)

	// UpdateRemindAt 更新提醒时间
	// 仅当任务 remind_at 仍为 expectedRemindAt 时生效（CAS 防与 Snooze 竞态）；
	// 返回是否实际变更（false 表示提醒已被其他路径改期/删除）
	UpdateRemindAt(ctx context.Context, taskId int64, expectedRemindAt time.Time, remindAt string) (bool, error)

	// SoftDeleteByProjectId 软删除指定项目下的所有任务（级联删除用）
	SoftDeleteByProjectId(ctx context.Context, userId int64, projectId int64) error

	// RemoveTagFromTasks 从所有任务中移除指定标签引用（标签删除时级联清理用）
	// 同时推进任务 updated_at，保证清理结果可被增量同步发现
	RemoveTagFromTasks(ctx context.Context, userId int64, tagId int64) error

	// RestoreByProjectId 恢复指定项目下的所有任务（级联恢复用）
	RestoreByProjectId(ctx context.Context, userId int64, projectId int64) error

	// ArchiveByProjectId 归档指定项目下的所有任务（级联归档用）
	ArchiveByProjectId(ctx context.Context, userId int64, projectId int64) error

	// UnarchiveByProjectId 取消归档指定项目下的所有任务（级联取消归档用）
	UnarchiveByProjectId(ctx context.Context, userId int64, projectId int64) error
}
