package project

import (
	"context"
	"errors"
	"naotodoserver/domain/project/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// RegistDomainImpl regist project domain impl
func RegistDomainImpl(projectDomain service.ProjectDomain) ProjectApp {
	once.Do(func() {
		App = &projectAppImpl{projectDomain: projectDomain}
	})
	return App
}

/*
 * Create project
 */
func (p *projectAppImpl) Create(
	ctx context.Context,
	req *types.CreateProjectReq,
) (*types.CreateProjectRes, error) {
	// 1. 获取 UserId
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 请求体转换实体
	e := CreateReq2Entity(req)
	e.UserId = userId
	// 3. 调用域函数 - 创建清单
	e, err := p.projectDomain.Create(ctx, e)
	if err != nil {
		return nil, err
	}
	// 4. 实体转换响应体
	res := Entity2CreateRes(e)
	// 5. 返回结果
	return res, nil
}

/*
 * Get project
 */
func (p *projectAppImpl) Get(
	ctx context.Context,
	req *types.GetProjectReq,
) (*types.GetProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 调用域函数 - 获取清单
	e, err := p.projectDomain.GetById(ctx, userId, projectId)
	if err != nil {
		return nil, err
	}
	// 4. 实体转换响应体并返回
	return Entity2GetRes(e), nil
}

/*
 * Update project
 */
func (p *projectAppImpl) Update(
	ctx context.Context,
	req *types.UpdateProjectReq,
) (*types.UpdateProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 请求体转换实体
	updateEntity := UpdateReq2Entity(req)
	// 4. 调用域函数 - 更新清单
	err = p.projectDomain.Update(ctx, userId, projectId, updateEntity)
	if err != nil {
		return nil, err
	}
	// 5. 返回结果
	return &types.UpdateProjectRes{ProjectId: req.ProjectId}, nil
}

/*
 * Delete project
 */
func (p *projectAppImpl) Delete(
	ctx context.Context,
	req *types.DeleteProjectReq,
) (*types.DeleteProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 调用域函数 - 删除清单
	err = p.projectDomain.Delete(ctx, userId, projectId)
	if err != nil {
		return nil, err
	}
	// 4. 返回结果
	return &types.DeleteProjectRes{ProjectId: req.ProjectId}, nil
}

/*
 * Restore project
 */
func (p *projectAppImpl) Restore(
	ctx context.Context,
	req *types.RestoreProjectReq,
) (*types.RestoreProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 调用域函数 - 恢复清单
	err = p.projectDomain.Restore(ctx, userId, projectId)
	if err != nil {
		return nil, err
	}
	// 4. 返回结果
	return &types.RestoreProjectRes{ProjectId: req.ProjectId}, nil
}

/*
 * Hard delete project
 */
func (p *projectAppImpl) HardDelete(
	ctx context.Context,
	req *types.HardDeleteProjectReq,
) (*types.HardDeleteProjectRes, error) {
	panic("unimplemented")
}

/*
 * Archive project
 */
func (p *projectAppImpl) Archive(
	ctx context.Context,
	req *types.ArchiveProjectReq,
) (*types.ArchiveProjectRes, error) {
	// 1. 获取 UserId
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 调用域函数 - 归档清单
	err = p.projectDomain.Archive(ctx, userId, projectId)
	if err != nil {
		return nil, err
	}
	// 4. 返回结果
	return &types.ArchiveProjectRes{ProjectId: req.ProjectId}, nil
}

/*
 * Unarchive project
 */
func (p *projectAppImpl) Unarchive(
	ctx context.Context,
	req *types.UnarchiveProjectReq,
) (*types.UnarchiveProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 获取清单 ID
	projectId, err := strconv.ParseInt(req.ProjectId, 10, 64)
	if err != nil {
		return nil, errors.New("参数错误 - 清单 ID 格式错误")
	}
	// 3. 调用域函数 - 取消归档清单
	err = p.projectDomain.Unarchive(ctx, userId, projectId)
	if err != nil {
		return nil, err
	}
	// 4. 返回结果
	return &types.UnarchiveProjectRes{ProjectId: req.ProjectId}, nil
}

/*
 * Update project preference
 */
func (p *projectAppImpl) UpdatePreference(
	ctx context.Context,
	req *types.UpdateProjectPreferenceReq,
) (*types.UpdateProjectPreferenceRes, error) {
	panic("unimplemented")
}

/*
 * List project
 */
func (p *projectAppImpl) List(ctx context.Context) (types.ListProjectRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId == 0 {
		return nil, errors.New("参数错误 - 用户 ID 不能为空")
	}
	// 2. 调用域函数 - 获取清单列表
	eList, err := p.projectDomain.GetByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 3. 转换实体
	res := Entities2ListRes(eList)
	// 3. 实体转换响应体并返回
	return res, nil
}
