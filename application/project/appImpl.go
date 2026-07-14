package project

import (
	"context"
	"errors"
	"fmt"
	"naotodoserver/application/idutil"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
)

// NewProjectApp 创建任务清单应用层实例
func NewProjectApp(
	projectDomain service.ProjectDomain,
	repo repositories.Project,
	preferenceRepo repositories.ProjectPreference,
) ProjectApp {
	impl := &projectAppImpl{
		projectDomain:  projectDomain,
		repo:           repo,
		preferenceRepo: preferenceRepo,
	}
	return impl
}

// 获取任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return *types.GetProjectRes 获取任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Get(
	ctx context.Context,
	projectId string,
) (*types.GetProjectRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, domerr.ErrInvalidUserID
	}
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
// @param createProjectReq 创建任务清单请求体
// @return *types.CreateProjectRes 创建任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Create(
	ctx context.Context,
	createProjectReq *types.CreateProjectReq,
) (*types.CreateProjectRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, domerr.ErrInvalidUserID
	}
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
	createProjectRes := (*types.CreateProjectRes)(ProjectEntityToGetRes(projectEntity))
	// 返回结果
	return createProjectRes, nil
}

// 更新任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @param updateProjectReq 更新任务清单请求体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Update(
	ctx context.Context,
	projectId string,
	updateProjectReq *types.UpdateProjectReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
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
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Delete(
	ctx context.Context,
	projectId string,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	// 删除清单
	err = app.projectDomain.Delete(ctx, userId, projectIdInt64)
	if err != nil {
		return errors.New("删除任务清单失败")
	}
	// 返回结果
	return nil
}

// 恢复任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Restore(
	ctx context.Context,
	projectId string,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	// 恢复清单
	err = app.projectDomain.Restore(ctx, userId, projectIdInt64)
	if err != nil {
		return err
	}
	// 返回结果
	return nil
}

// 硬删除任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
// DeleteDeactivatedProjects 删除已注销的任务清单（供定时任务调用）
func (app *projectAppImpl) DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) error {
	_, err := app.repo.DeleteDeactivatedProjects(ctx, dayOffset)
	return err
}

func (app *projectAppImpl) HardDelete(
	ctx context.Context,
	projectId string,
) error {
	panic("unimplemented")
}

// 归档任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Archive(
	ctx context.Context,
	projectId string,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	return app.repo.Archive(ctx, userId, projectIdInt64)
}

// 取消归档任务清单
// @param ctx 上下文
// @param projectId 任务清单 ID
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) Unarchive(
	ctx context.Context,
	projectId string,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
	// 获取清单 ID
	projectIdInt64, err := idutil.ParseID(projectId)
	if err != nil {
		return domerr.ErrInvalidProjectID
	}
	return app.repo.Unarchive(ctx, userId, projectIdInt64)
}

// 获取用户任务清单列表
// @param ctx 上下文
// @return types.ListProjectRes 任务清单响应体列表
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) List(ctx context.Context) (types.ListProjectRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, domerr.ErrInvalidUserID
	}
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
// @param req 批量更新任务清单请求体
// @return *types.BatchUpdateProjectRes 批量更新任务清单响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) BatchUpdate(
	ctx context.Context,
	req *types.BatchUpdateProjectReq,
) (*types.BatchUpdateProjectRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, domerr.ErrInvalidUserID
	}
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
	return &types.BatchUpdateProjectRes{
		UpdatedCount: int64(len(projectResList)),
		Projects:     projectResList,
	}, nil
}

// 获取任务清单偏好
// @param ctx 上下文
// @param req 获取任务清单偏好请求体
// @return *types.ProjectPreferenceRes 任务清单偏好响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) GetPreference(
	ctx context.Context,
	projectId string,
) (*types.GetProjectPreferenceRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, domerr.ErrInvalidUserID
	}
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
// @param projectId 任务清单 ID
// @param req 更新任务清单偏好请求体
// @return *types.UpdateProjectPreferenceRes 更新任务清单偏好响应体
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) SavePreference(
	ctx context.Context,
	projectId string,
	updateProjectPreferenceReq *types.UpdateProjectPreferenceReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return domerr.ErrInvalidUserID
	}
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
