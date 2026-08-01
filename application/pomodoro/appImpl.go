package pomodoro

import (
	"context"
	"fmt"

	"naotodoserver/application/idutil"
	"naotodoserver/application/pomodoro/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/service"
	iCtx "naotodoserver/infrastructure/context"
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
	req *dto.CreatePomodoroRecordReq,
) (*dto.CreatePomodoroRecordRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	vo, err := CreatePomodoroRecordReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	entity, err := app.pomodoroDomain.CreatePomodoroRecord(ctx, userId, vo)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.Create: %w", err)
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
	req *dto.GetPomodoroRecordReq,
) (*dto.GetPomodoroRecordRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	id, err := idutil.ParseID(req.Id)
	if err != nil {
		return nil, domerr.ErrInvalidID
	}
	entity, err := app.pomodoroRecordRepo.GetById(ctx, userId, id)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.Get: %w", err)
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
	req *dto.ListPomodoroRecordReq,
) ([]*dto.GetPomodoroRecordRes, int64, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, 0, domerr.ErrInvalidUserID
	}
	q := ListPomodoroRecordReqToQueryVO(userId, req)
	entities, total, err := app.pomodoroRecordRepo.List(ctx, userId, q)
	if err != nil {
		return nil, 0, fmt.Errorf("pomodoroRecord.List: %w", err)
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
	req *dto.CreatePomodoroReq,
) (*dto.CreatePomodoroRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	vo, err := CreatePomodoroReqToVO(userId, req)
	if err != nil {
		return nil, err
	}
	entity, err := app.pomodoroDomain.CreatePomodoro(ctx, userId, vo)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.CreatePomodoro: %w", err)
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
	req *dto.GetPomodoroReq,
) (*dto.PomodoroRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	id, err := idutil.ParseID(req.Id)
	if err != nil {
		return nil, domerr.ErrInvalidPomodoroID
	}
	entity, err := app.pomodoroRepo.GetById(ctx, userId, id)
	if err != nil {
		return nil, fmt.Errorf("pomodoro.GetPomodoro: %w", err)
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
	req *dto.UpdatePomodoroReq,
) error {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	idInt64, err := idutil.ParseID(id)
	if err != nil {
		return domerr.ErrInvalidPomodoroID
	}
	vo, err := UpdatePomodoroReqToVO(req)
	if err != nil {
		return err
	}
	_, err = app.pomodoroDomain.UpdatePomodoro(ctx, userId, idInt64, vo)
	if err != nil {
		return fmt.Errorf("pomodoro.Update: %w", err)
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
		return domerr.ErrInvalidUserID
	}
	idInt64, err := idutil.ParseID(id)
	if err != nil {
		return domerr.ErrInvalidPomodoroID
	}
	if err := app.pomodoroRepo.Delete(ctx, userId, idInt64); err != nil {
		return fmt.Errorf("pomodoro.Delete: %w", err)
	}
	return nil
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
		return domerr.ErrInvalidUserID
	}
	idInt64, err := idutil.ParseID(id)
	if err != nil {
		return domerr.ErrInvalidPomodoroID
	}
	if err := app.pomodoroRepo.Archive(ctx, userId, idInt64); err != nil {
		return fmt.Errorf("pomodoro.Archive: %w", err)
	}
	return nil
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
		return domerr.ErrInvalidUserID
	}
	idInt64, err := idutil.ParseID(id)
	if err != nil {
		return domerr.ErrInvalidPomodoroID
	}
	if err := app.pomodoroRepo.Unarchive(ctx, userId, idInt64); err != nil {
		return fmt.Errorf("pomodoro.Unarchive: %w", err)
	}
	return nil
}

// ListPomodoro 获取常用番茄工作列表
// @param ctx 上下文
// @param req 获取常用番茄工作列表请求
// @return 常用番茄工作列表响应
// @return total 总记录数
// @return error 错误
func (app *PomodoroAppImpl) ListPomodoro(
	ctx context.Context,
	req *dto.ListPomodoroReq,
) (dto.ListPomodoroRes, int64, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, 0, domerr.ErrInvalidUserID
	}
	q := ListPomodoroReqToQueryVO(userId, req)
	entities, total, err := app.pomodoroRepo.List(ctx, userId, q)
	if err != nil {
		return nil, 0, fmt.Errorf("pomodoro.List: %w", err)
	}
	resList := PomodoroEntitiesToGetReses(entities)
	return resList, total, nil
}
