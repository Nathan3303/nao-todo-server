package project

import (
	"context"
	"naotodoserver/domain/project/service"
	"naotodoserver/interfaces/types"
	"sync"
)

// 任务清单应用接口
type ProjectApp interface {
	// 获取任务清单
	Get(ctx context.Context, projectId string) (*types.GetProjectRes, error)

	// 创建任务清单
	Create(
		ctx context.Context,
		createProjectReq *types.CreateProjectReq,
	) (*types.CreateProjectRes, error)

	// 更新任务清单
	Update(ctx context.Context, projectId string, updateProjectReq *types.UpdateProjectReq) error

	// 删除任务清单
	Delete(ctx context.Context, projectId string) error

	// 恢复任务清单
	Restore(ctx context.Context, projectId string) error

	// 硬删除任务清单
	HardDelete(ctx context.Context, projectId string) error

	// 归档任务清单
	Archive(ctx context.Context, projectId string) error

	// 取消归档任务清单
	Unarchive(ctx context.Context, projectId string) error

	// 获取任务清单列表
	List(ctx context.Context) (types.ListProjectRes, error)

	// 获取任务清单偏好
	GetPreference(ctx context.Context, projectId string) (*types.GetProjectPreferenceRes, error)

	// 保存任务清单偏好
	SavePreference(
		ctx context.Context,
		projectId string,
		updateProjectPreferenceReq *types.UpdateProjectPreferenceReq,
	) error
}

// 任务清单应用实现
type projectAppImpl struct {
	projectDomain service.ProjectDomain
}

// 任务清单应用单例
var (
	App  *projectAppImpl
	once sync.Once
)
