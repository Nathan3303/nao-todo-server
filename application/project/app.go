package project

import (
	"context"
	"naotodoserver/application/project/dto"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/service"
	taskRepo "naotodoserver/domain/task/repositories"
	domaintypes "naotodoserver/domain/types"
)

// ProjectApp 任务清单应用接口
type ProjectApp interface {
	// 获取任务清单
	Get(ctx context.Context, userId int64, projectId string) (*dto.GetProjectRes, error)

	// 创建任务清单
	Create(
		ctx context.Context,
		userId int64,
		createProjectReq *dto.CreateProjectReq,
	) (*dto.CreateProjectRes, error)

	// 更新任务清单
	Update(
		ctx context.Context, userId int64, projectId string, updateProjectReq *dto.UpdateProjectReq,
	) error

	// 删除任务清单
	Delete(ctx context.Context, userId int64, projectId string) error

	// 恢复任务清单
	Restore(ctx context.Context, userId int64, projectId string) error

	// 归档任务清单
	Archive(ctx context.Context, userId int64, projectId string) error

	// 取消归档任务清单
	Unarchive(ctx context.Context, userId int64, projectId string) error

	// 获取任务清单列表
	List(ctx context.Context, userId int64) (dto.ListProjectRes, error)

	// 批量更新任务清单
	BatchUpdate(
		ctx context.Context,
		userId int64,
		req *dto.BatchUpdateProjectReq,
	) (*dto.BatchUpdateProjectRes, error)

	// 获取任务清单偏好
	GetPreference(
		ctx context.Context, userId int64, projectId string,
	) (*dto.GetProjectPreferenceRes, error)

	// 保存任务清单偏好
	SavePreference(
		ctx context.Context,
		userId int64,
		projectId string,
		updateProjectPreferenceReq *dto.UpdateProjectPreferenceReq,
	) error

	// 删除已注销的任务清单（供定时任务调用）
	DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) error
}

// projectAppImpl 任务清单应用实现
type projectAppImpl struct {
	projectDomain  service.ProjectDomain
        txManager      domaintypes.TxManager
	repo           repositories.Project
	preferenceRepo repositories.ProjectPreference
	taskRepo       taskRepo.Task           // 用于级联操作
}
