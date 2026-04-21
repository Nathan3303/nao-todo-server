package project

import (
	"context"
	"errors"
	"naotodoserver/domain/project/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// 注册任务清单应用实现
func RegistDomainImpl(projectDomain service.ProjectDomain) ProjectApp {
	once.Do(func() {
		App = &projectAppImpl{projectDomain: projectDomain}
	})
	return App
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
		return nil, errors.New("用户 ID 不能为空")
	}
	// 转换清单 ID 为 int64 类型
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return nil, errors.New("清单 ID 格式错误")
	}
	// 获取清单
	projectEntity, err := app.projectDomain.GetById(ctx, userId, projectIdInt64)
	if err != nil {
		return nil, err
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
		return nil, errors.New("用户 ID 不能为空")
	}
	// 请求体转换值对象
	createProjectValueObject, err := CreateProjectReqToValueObject(userId, createProjectReq)
	if err != nil {
		return nil, err
	}
	// 创建任务清单
	projectEntity, err := app.projectDomain.Create(ctx, createProjectValueObject)
	if err != nil {
		return nil, err
	}
	// 实体转换响应体
	createProjectRes := ProjectEntityToCreateRes(projectEntity)
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
		return errors.New("用户 ID 不能为空")
	}
	// 转换清单 ID 为 int64 类型
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
	}
	// 请求体转换实体
	updateProjectValueObject, err := UpdateProjectReqToValueObject(updateProjectReq)
	if err != nil {
		return errors.New("更新任务清单请求体格式错误")
	}
	// 调用域函数 - 更新清单
	err = app.projectDomain.Update(ctx, userId, projectIdInt64, updateProjectValueObject)
	if err != nil {
		return errors.New("更新任务清单失败")
	}
	// 返回结果
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
		return errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
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
		return errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
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
		return errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
	}
	// 归档清单
	err = app.projectDomain.Archive(ctx, userId, projectIdInt64)
	if err != nil {
		return err
	}
	// 返回结果
	return nil
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
		return errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
	}
	// 取消归档清单
	err = app.projectDomain.Unarchive(ctx, userId, projectIdInt64)
	if err != nil {
		return err
	}
	// 返回结果
	return nil
}

// 获取用户任务清单列表
// @param ctx 上下文
// @return types.ListProjectRes 任务清单响应体列表
// @return error 验证失败返回错误，否则返回 nil
func (app *projectAppImpl) List(ctx context.Context) (types.ListProjectRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("用户 ID 不能为空")
	}
	// 获取清单列表
	projectEntities, err := app.projectDomain.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 转换实体
	getResList := EntitiesToGetResList(projectEntities)
	// 实体转换响应体并返回
	return getResList, nil
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
		return nil, errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return nil, errors.New("清单 ID 格式错误")
	}
	// 获取清单偏好
	projectPreferenceEntity, err := app.projectDomain.GetPreference(
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
		return errors.New("用户 ID 不能为空")
	}
	// 获取清单 ID
	projectIdInt64, err := strconv.ParseInt(projectId, 10, 64)
	if err != nil {
		return errors.New("清单 ID 格式错误")
	}
	// 请求体转换实体
	saveProjectPreferenceValueObject, err := UpdateProjectPreferenceReqToValueObject(
		updateProjectPreferenceReq,
	)
	if err != nil {
		return err
	}
	// 更新清单偏好
	err = app.projectDomain.SavePreference(
		ctx,
		userId,
		projectIdInt64,
		saveProjectPreferenceValueObject,
	)
	if err != nil {
		return err
	}
	// 返回结果
	return nil
}
