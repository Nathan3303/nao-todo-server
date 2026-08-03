package project

import (
	"context"
	"errors"
	"fmt"

	"naotodoserver/application/idutil"
	"naotodoserver/application/project/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/service"
	taskRepo "naotodoserver/domain/task/repositories"
        domaintypes "naotodoserver/domain/types"
)

// NewProjectApp 创建任务清单应用层实例
func NewProjectApp(
	projectDomain service.ProjectDomain,
        txManager domaintypes.TxManager,
	repo repositories.Project,
	preferenceRepo repositories.ProjectPreference,
	taskRepo taskRepo.Task,
) ProjectApp {
	impl := &projectAppImpl{
		projectDomain:  projectDomain,
                txManager:      txManager,
		repo:           repo,
		preferenceRepo: preferenceRepo,
		taskRepo:       taskRepo,
	}
	return impl
}

// 获取任务清单
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @return *dto.GetProjectRes 获取任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Get(
	ctx context.Context,
	userId int64,
	projectId string,
) (*dto.GetProjectRes, error) {
	// 转换清单 ID 为 int64 类型
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return nil, domerr.ErrInvalidProjectID
	}
	// 获取清单
	projectEntity, err := app.repo.GetById(ctx, userId, projectIdInt64)
	if err != nil {
		return nil, fmt.Errorf("project.Get: %w", err)
	}
	// 实体转换响应体并返回
	return ProjectEntityToGetRes(projectEntity), nil
}

// 创建任务清单
// @param ctx 上下文
// @param userId 用户 ID
// @param createProjectReq 创建任务清单请求体
// @return *dto.CreateProjectRes 创建任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Create(
	ctx context.Context,
	userId int64,
	createProjectReq *dto.CreateProjectReq,
) (*dto.CreateProjectRes, error) {
	// 请求体转换值对象
	createProjectValueObject, err := CreateProjectReqToValueObject(userId, createProjectReq)
	if err != nil {
		return nil, err
	}
	// 创建任务清单
	projectEntity, err := app.projectDomain.Create(ctx, createProjectValueObject)
	if err != nil {
		return nil, fmt.Errorf("project.Create: %w", err)
	}
	// 实体转换响应体
	createProjectRes := (*dto.CreateProjectRes)(ProjectEntityToGetRes(projectEntity))
	// 返回结果
	return createProjectRes, nil
}

// 更新任务清单
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @param updateProjectReq 更新任务清单请求体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Update(
	ctx context.Context,
	userId int64,
	projectId string,
	updateProjectReq *dto.UpdateProjectReq,
) error {
	// 转换清单 ID 为 int64 类型
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	// 请求体转换实体
	updateProjectValueObject, err := UpdateProjectReqToValueObject(updateProjectReq)
	if err != nil {
		return errors.New("更新任务清单请求体格式错误: " + err.Error())
	}
	if err := app.repo.Update(ctx, userId, projectIdInt64, updateProjectValueObject); err != nil {
		return fmt.Errorf("project.Update: %w", err)
	}
	return nil
}

// 删除任务清单
// 同步级联删除 Project 下所有 Task
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Delete(
	ctx context.Context,
	userId int64,
	projectId string,
) error {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}

        err = app.txManager.Do(ctx, func(ctx context.Context) error {
                // 1. 删除 Project（领域服务负责 Project 自身 + Preference）
                if err := app.projectDomain.Delete(ctx, userId, projectIdInt64); err != nil {
                        return errors.New("删除任务清单失败")
                }
                // 2. 级联软删除归属于该 Project 的所有 Task
                err = app.taskRepo.SoftDeleteByProjectId(ctx, userId, projectIdInt64)
                if err != nil {
                        return err
                }
                return nil
        })
        if err != nil {
                return err
        }
        return nil
}

// 恢复任务清单
// 同步级联恢复 Project 下所有 Task
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId string,
) error {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}

        err = app.txManager.Do(ctx, func(ctx context.Context) error {
                // 1. 恢复 Project
                if err := app.projectDomain.Restore(ctx, userId, projectIdInt64); err != nil {
                        return err
                }
                // 2. 级联恢复归属于该 Project 的所有 Task
                if err := app.taskRepo.RestoreByProjectId(ctx, userId, projectIdInt64); err != nil {
                        return err
                }
                return nil
        })
        if err != nil {
                return err
        }
        return nil
}

// DeleteDeactivatedProjects 删除已注销的任务清单（供定时任务调用）
func (app *projectAppImpl) DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) error {
	_, err := app.repo.DeleteDeactivatedProjects(ctx, dayOffset)
	return err
}

// 归档任务清单
// 同步级联归档 Project 下所有 Task
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId string,
) error {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}

        err = app.txManager.Do(ctx, func(ctx context.Context) error {
                // 1. 归档 Project
                if err := app.projectDomain.Archive(ctx, userId, projectIdInt64); err != nil {
                        return err
                }
                // 2. 级联归档归属于该 Project 的所有 Task
                if err := app.taskRepo.ArchiveByProjectId(ctx, userId, projectIdInt64); err != nil {
                        return err
                }
                return nil
        })
        if err != nil {
                return err
        }
        return nil
}

// 取消归档任务清单
// 同步级联取消归档 Project 下所有 Task
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId string,
) error {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}

        err = app.txManager.Do(ctx, func(ctx context.Context) error {
                // 1. 取消归档 Project
                if err := app.projectDomain.Unarchive(ctx, userId, projectIdInt64); err != nil {
                        return err
                }
                // 2. 级联取消归档归属于该 Project 的所有 Task
                err = app.taskRepo.UnarchiveByProjectId(ctx, userId, projectIdInt64)
                if err != nil {
                        return err
                }
                return nil
        })
        if err != nil {
                return err
        }
        return nil
}

// 获取用户任务清单列表
// @param ctx 上下文
// @param userId 用户 ID
// @return dto.ListProjectRes 任务清单响应体列表
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) List(ctx context.Context, userId int64) (dto.ListProjectRes, error) {
	// 获取清单列表
	projectEntities, err := app.repo.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 转换实体
	return EntitiesToGetResList(projectEntities), nil
}

// 批量更新任务清单
// @param ctx 上下文
// @param userId 用户 ID
// @param req 批量更新任务清单请求体
// @return *dto.BatchUpdateProjectRes 批量更新任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	req *dto.BatchUpdateProjectReq,
) (*dto.BatchUpdateProjectRes, error) {
	// 请求体转换值对象
	batchVOs, err := BatchUpdateProjectReqToValueObjects(req)
	if err != nil {
		return nil, err
	}
	updatedEntities, err := app.repo.BatchUpdate(ctx, userId, batchVOs)
	if err != nil {
		return nil, err
	}
	projectResList := EntitiesToGetResList(updatedEntities)
	// 返回结果
	return &dto.BatchUpdateProjectRes{
		UpdatedCount: int64(len(projectResList)),
		Projects:     projectResList,
	}, nil
}

// 获取任务清单偏好
// @param ctx 上下文
// @param userId 用户 ID
// @param req 获取任务清单偏好请求体
// @return *dto.GetProjectPreferenceRes 任务清单偏好响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) GetPreference(
	ctx context.Context,
	userId int64,
	projectId string,
) (*dto.GetProjectPreferenceRes, error) {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return nil, domerr.ErrInvalidProjectID
	}
	// 获取清单偏好
	projectPreferenceEntity, err := app.preferenceRepo.Get(
		ctx,
		userId,
		projectIdInt64,
	)
	if err != nil {
		return nil, err
	}
	// 返回
	return ProjectPreferenceEntityToGetRes(projectPreferenceEntity), nil
}

// 保存任务清单偏好
// @param ctx 上下文
// @param userId 用户 ID
// @param projectId 任务清单 ID
// @param req 更新任务清单偏好请求体
// @return *dto.UpdateProjectPreferenceRes 更新任务清单偏好响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) SavePreference(
	ctx context.Context,
	userId int64,
	projectId string,
	updateProjectPreferenceReq *dto.UpdateProjectPreferenceReq,
) error {
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	// 请求体转换实体
	saveProjectPreferenceValueObject, err := UpdateProjectPreferenceReqToValueObject(
		updateProjectPreferenceReq,
	)
	if err != nil {
		return err
	}
	// 更新清单偏好
	return app.preferenceRepo.Save(
		ctx,
		userId,
		projectIdInt64,
		saveProjectPreferenceValueObject,
	)
}
