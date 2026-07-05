package pomodoro

import (
	"context"
	"errors"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewPomodoroApp 创建专注应用应用层实例
func NewPomodoroApp(
	pomodoroDomain service.PomodoroDomain,
	pomodoroRecordRepo repositories.PomodoroRecord,
	pomodoroRepo repositories.Pomodoro,
) PomodoroApp {
	return &PomodoroAppImpl{
		pomodoroDomain:     pomodoroDomain,
		pomodoroRecordRepo: pomodoroRecordRepo,
		pomodoroRepo:       pomodoroRepo,
	}
}

// --- PomodoroRecord ---

// Create 创建专注记录
// @param ctx 上下文
// @param req 创建专注记录请求
// @return 创建专注记录响应
// @return error 错误
func (app *PomodoroAppImpl) Create(
	ctx context.Context,
	req *types.CreatePomodoroRecordReq,
) (*types.CreatePomodoroRecordRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	vo, err := CreatePomodoroRecordReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	entity, err := app.pomodoroDomain.CreatePomodoroRecord(ctx, userId, vo)
	if err != nil {
		return nil, err
	}
	return PomodoroRecordEntityToCreateRes(entity), nil
}

// Get 获取专注记录
// @param ctx 上下文
// @param req 获取专注记录请求
// @return 获取专注记录响应
// @return error 错误
func (app *PomodoroAppImpl) Get(
	ctx context.Context,
	req *types.GetPomodoroRecordReq,
) (*types.GetPomodoroRecordRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, errors.New("专注记录 ID 无效")
	}
	entity, err := app.pomodoroRecordRepo.GetById(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	return PomodoroRecordEntityToGetRes(entity), nil
}

// List 获取专注记录列表
// @param ctx 上下文
// @param req 获取专注记录列表请求
// @return 获取专注记录列表响应
// @return error 错误
func (app *PomodoroAppImpl) List(
	ctx context.Context,
	req *types.ListPomodoroRecordReq,
) ([]*types.GetPomodoroRecordRes, int64, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, 0, errors.New("用户 ID 无效")
	}
	q := ListPomodoroRecordReqToQueryVO(userId, req)
	entities, total, err := app.pomodoroRecordRepo.List(ctx, userId, q)
	if err != nil {
		return nil, 0, err
	}
	resList := PomodoroRecordEntitiesToGetReses(entities)
	return resList, total, nil
}

// --- Pomodoro ---

// CreatePomodoro 创建常用番茄工作
// @param ctx 上下文
// @param req 创建常用番茄工作请求
// @return 创建常用番茄工作响应
// @return error 错误
func (app *PomodoroAppImpl) CreatePomodoro(
	ctx context.Context,
	req *types.CreatePomodoroReq,
) (*types.CreatePomodoroRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	vo, err := CreatePomodoroReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	entity, err := app.pomodoroDomain.CreatePomodoro(ctx, userId, vo)
	if err != nil {
		return nil, err
	}
	return PomodoroEntityToCreateRes(entity), nil
}

// GetPomodoro 获取常用番茄工作
// @param ctx 上下文
// @param req 获取常用番茄工作请求
// @return 常用番茄工作响应
// @return error 错误
func (app *PomodoroAppImpl) GetPomodoro(
	ctx context.Context,
	req *types.GetPomodoroReq,
) (*types.PomodoroRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, errors.New("常用番茄工作 ID 无效")
	}
	entity, err := app.pomodoroRepo.GetById(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	return PomodoroEntityToGetRes(entity), nil
}

// UpdatePomodoro 更新常用番茄工作（PATCH 语义）
// @param ctx 上下文
// @param id 常用番茄工作 ID
// @param req 更新常用番茄工作请求
// @return error 错误
func (app *PomodoroAppImpl) UpdatePomodoro(
	ctx context.Context,
	id string,
	req *types.UpdatePomodoroReq,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("常用番茄工作 ID 无效")
	}
	vo, err := UpdatePomodoroReqToVO(req)
	if err != nil {
		return err
	}
	_, err = app.pomodoroDomain.UpdatePomodoro(ctx, userId, idInt64, vo)
	if err != nil {
		return err
	}
	return nil
}

// DeletePomodoro 删除常用番茄工作（软删除）
// @param ctx 上下文
// @param id 常用番茄工作 ID
// @return error 错误
func (app *PomodoroAppImpl) DeletePomodoro(
	ctx context.Context,
	id string,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("常用番茄工作 ID 无效")
	}
	return app.pomodoroRepo.Delete(ctx, userId, idInt64)
}

// ArchivePomodoro 归档常用番茄工作
// @param ctx 上下文
// @param id 常用番茄工作 ID
// @return error 错误
func (app *PomodoroAppImpl) ArchivePomodoro(
	ctx context.Context,
	id string,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("常用番茄工作 ID 无效")
	}
	return app.pomodoroRepo.Archive(ctx, userId, idInt64)
}

// UnarchivePomodoro 取消归档常用番茄工作
// @param ctx 上下文
// @param id 常用番茄工作 ID
// @return error 错误
func (app *PomodoroAppImpl) UnarchivePomodoro(
	ctx context.Context,
	id string,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	idInt64, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return errors.New("常用番茄工作 ID 无效")
	}
	return app.pomodoroRepo.Unarchive(ctx, userId, idInt64)
}

// ListPomodoro 获取常用番茄工作列表
// @param ctx 上下文
// @param req 获取常用番茄工作列表请求
// @return 常用番茄工作列表响应
// @return total 总记录数
// @return error 错误
func (app *PomodoroAppImpl) ListPomodoro(
	ctx context.Context,
	req *types.ListPomodoroReq,
) (types.ListPomodoroRes, int64, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, 0, errors.New("用户 ID 无效")
	}
	q := ListPomodoroReqToQueryVO(userId, req)
	entities, total, err := app.pomodoroRepo.List(ctx, userId, q)
	if err != nil {
		return nil, 0, err
	}
	resList := PomodoroEntitiesToGetReses(entities)
	return resList, total, nil
}
