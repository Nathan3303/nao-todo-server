package pomodoro

import (
	"context"
	"errors"
	"naotodoserver/domain/pomodoro/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewPomodoroApp 创建专注应用应用层实例
func NewPomodoroApp(pomodoroDomain service.PomodoroDomain) PomodoroApp {
	return &PomodoroAppImpl{pomodoroDomain: pomodoroDomain}
}

// Create 创建专注记录
// @param ctx 上下文
// @param req 创建专注记录请求
// @return 创建专注记录响应
// @return error 错误
func (app *PomodoroAppImpl) Create(
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
	entity, err := app.pomodoroDomain.Create(ctx, userId, vo)
	if err != nil {
		return nil, err
	}
	return PomodoroEntityToCreateRes(entity), nil
}

// Get 获取专注记录
// @param ctx 上下文
// @param req 获取专注记录请求
// @return 获取专注记录响应
// @return error 错误
func (app *PomodoroAppImpl) Get(
	ctx context.Context,
	req *types.GetPomodoroReq,
) (*types.GetPomodoroRes, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	id, err := strconv.ParseInt(req.Id, 10, 64)
	if err != nil {
		return nil, errors.New("专注记录 ID 无效")
	}
	entity, err := app.pomodoroDomain.GetById(ctx, userId, id)
	if err != nil {
		return nil, err
	}
	return PomodoroEntityToGetRes(entity), nil
}

// List 获取专注记录列表
// @param ctx 上下文
// @param req 获取专注记录列表请求
// @return 获取专注记录列表响应
// @return error 错误
func (app *PomodoroAppImpl) List(
	ctx context.Context,
	req *types.ListPomodoroReq,
) ([]*types.GetPomodoroRes, int64, error) {
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, 0, errors.New("用户 ID 无效")
	}
	var taskId int64
	if req.TaskId != "" {
		var err error
		taskId, err = strconv.ParseInt(req.TaskId, 10, 64)
		if err != nil {
			taskId = 0
		}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}
	entities, total, err := app.pomodoroDomain.List(
		ctx,
		userId,
		req.SessionId,
		req.StartTime,
		req.EndTime,
		taskId,
		req.TaskName,
		req.Type,
		page,
		limit,
		req.Sort,
	)
	if err != nil {
		return nil, 0, err
	}
	resList := PomodoroEntitiesToGetReses(entities)
	return resList, total, nil
}
