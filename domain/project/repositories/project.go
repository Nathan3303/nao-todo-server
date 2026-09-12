package repositories

import (
	"context"
	"time"

	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/domain/types"
)

// Project 任务清单仓库接口
type Project interface {
	// 创建任务清单
	Create(
		ctx context.Context,
		createProjectValueObject *valueobjects.CreateProject,
	) (*entities.Project, error)

	// Upsert 幂等写入：客户端指定 id 时创建/覆盖（LWW 判定 + create 冲突检测）
	// created=true 表示本次为新建（调用方需初始化偏好等附属记录）
	Upsert(
		ctx context.Context,
		userId int64,
		createProjectValueObject *valueobjects.CreateProject,
	) (*entities.Project, bool, error)

	// 根据用户ID和任务清单ID获取任务清单
	GetById(ctx context.Context, userId int64, projectId int64) (*entities.Project, error)

	// ListSync 增量同步列表：包含软删墓碑，(updated_at, id) keyset 游标稳定排序分页（不缓存）
	ListSync(
		ctx context.Context,
		userId int64,
		cursor time.Time,
		cursorID int64,
		limit int,
	) ([]*entities.Project, error)

	// 更新任务清单
	Update(
		ctx context.Context,
		userId int64,
		projectId int64,
		updateProjectValueObject *valueobjects.UpdateProject,
	) error

	// 更新任务清单状态（归档/停用时间字段）
	UpdateState(
		ctx context.Context,
		userId int64,
		projectId int64,
		archivedAt types.NullableTime,
		deactivedAt types.NullableTime,
	) error

	// 根据用户ID获取任务清单列表
	GetByUserId(ctx context.Context, userId int64) ([]*entities.Project, error)

	// 获取最大排序 ID
	GetMaxSortId(ctx context.Context, userId int64) uint16

	// 批量更新任务清单
	BatchUpdate(
		ctx context.Context,
		userId int64,
		batchUpdateProjects []*valueobjects.BatchUpdateProject,
	) ([]*entities.Project, error)

	// 删除已注销的任务清单
	DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) (int64, error)

	// AdjustTaskCount 调整项目任务计数（E4/E5；直写列 + bump updated_at；隐式桶无 projects 行为无害 no-op）
	AdjustTaskCount(ctx context.Context, userId, projectId int64, delta int) error

	// RecountTaskCount 批量重算项目任务计数（E7 级联删/恢复：写最终值，不逐事件；口径含子任务/含归档/含放弃、不含已删除）
	RecountTaskCount(ctx context.Context, userId, projectId int64) error
}
