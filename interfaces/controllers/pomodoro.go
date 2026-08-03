package controllers

import (
	iCtx "naotodoserver/infrastructure/context"
	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// --- Pomodoro Record ---

type PomodoroController struct {
	pomodoroApp pomodoroApp.PomodoroApp
}

func NewPomodoroController(app pomodoroApp.PomodoroApp) *PomodoroController {
	return &PomodoroController{pomodoroApp: app}
}

// toCreatePomodoroRecordInput 将创建番茄工作记录请求转换为应用层入参
// @param req 创建番茄工作记录请求
// @return 应用层创建番茄工作记录入参
func toCreatePomodoroRecordInput(
	req types.CreatePomodoroRecordReq,
) *pomodoroDto.CreatePomodoroRecordReq {
	return &pomodoroDto.CreatePomodoroRecordReq{
		SessionId:   req.SessionId,
		PomodoroId:  req.PomodoroId,
		Type:        req.Type,
		TaskId:      req.TaskId,
		TaskName:    req.TaskName,
		Description: req.Description,
		StartAt:     req.StartAt,
		EndAt:       req.EndAt,
		Duration:    req.Duration,
		Note:        req.Note,
	}
}

// toCreatePomodoroRecordRes 将应用层创建番茄工作记录出参转换为创建番茄工作记录响应
// @param output 应用层创建番茄工作记录出参
// @return 创建番茄工作记录响应
func toCreatePomodoroRecordRes(
	output *pomodoroDto.CreatePomodoroRecordRes,
) *types.CreatePomodoroRecordRes {
	return &types.CreatePomodoroRecordRes{
		ResBase: types.ResBase{
			Id:        output.Id,
			CreatedAt: output.CreatedAt,
			UpdatedAt: output.UpdatedAt,
			DeletedAt: output.DeletedAt,
		},
		SessionId:   output.SessionId,
		PomodoroId:  output.PomodoroId,
		Type:        output.Type,
		TaskId:      output.TaskId,
		TaskName:    output.TaskName,
		Description: output.Description,
		StartAt:     output.StartAt,
		EndAt:       output.EndAt,
		Duration:    output.Duration,
		Note:        output.Note,
	}
}

// toGetPomodoroRecordInput 将获取番茄工作记录请求转换为应用层入参
// @param req 获取番茄工作记录请求
// @return 应用层获取番茄工作记录入参
func toGetPomodoroRecordInput(req types.GetPomodoroRecordReq) *pomodoroDto.GetPomodoroRecordReq {
	return &pomodoroDto.GetPomodoroRecordReq{
		Id: req.Id,
	}
}

// toGetPomodoroRecordRes 将应用层获取番茄工作记录出参转换为获取番茄工作记录响应
// @param output 应用层获取番茄工作记录出参
// @return 获取番茄工作记录响应
func toGetPomodoroRecordRes(output *pomodoroDto.GetPomodoroRecordRes) *types.GetPomodoroRecordRes {
	create := toCreatePomodoroRecordRes((*pomodoroDto.CreatePomodoroRecordRes)(output))
	return (*types.GetPomodoroRecordRes)(create)
}

// toGetPomodoroRecordReses 将应用层获取番茄工作记录列表出参转换为获取番茄工作记录列表响应
// @param output 应用层获取番茄工作记录列表出参
// @return 获取番茄工作记录列表响应
func toGetPomodoroRecordReses(
	output []*pomodoroDto.GetPomodoroRecordRes,
) []*types.GetPomodoroRecordRes {
	res := make([]*types.GetPomodoroRecordRes, 0, len(output))
	for _, item := range output {
		res = append(res, toGetPomodoroRecordRes(item))
	}
	return res
}

// toListPomodoroRecordInput 将获取番茄工作记录列表请求转换为应用层入参
// @param req 获取番茄工作记录列表请求
// @return 应用层获取番茄工作记录列表入参
func toListPomodoroRecordInput(req types.ListPomodoroRecordReq) *pomodoroDto.ListPomodoroRecordReq {
	return &pomodoroDto.ListPomodoroRecordReq{
		PomodoroId: req.PomodoroId,
		SessionId:  req.SessionId,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		TaskId:     req.TaskId,
		TaskName:   req.TaskName,
		Type:       req.Type,
		Page:       req.Page,
		Limit:      req.Limit,
		Sort:       req.Sort,
	}
}

// toCreatePomodoroInput 将创建常用番茄工作请求转换为应用层入参
// @param req 创建常用番茄工作请求
// @return 应用层创建常用番茄工作入参
func toCreatePomodoroInput(req types.CreatePomodoroReq) *pomodoroDto.CreatePomodoroReq {
	return &pomodoroDto.CreatePomodoroReq{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Duration:    req.Duration,
	}
}

// toGetPomodoroInput 将获取常用番茄工作请求转换为应用层入参
// @param req 获取常用番茄工作请求
// @return 应用层获取常用番茄工作入参
func toGetPomodoroInput(req types.GetPomodoroReq) *pomodoroDto.GetPomodoroReq {
	return &pomodoroDto.GetPomodoroReq{
		Id: req.Id,
	}
}

// toUpdatePomodoroInput 将更新常用番茄工作请求转换为应用层入参
// @param req 更新常用番茄工作请求
// @return 应用层更新常用番茄工作入参
func toUpdatePomodoroInput(req types.UpdatePomodoroReq) *pomodoroDto.UpdatePomodoroReq {
	return &pomodoroDto.UpdatePomodoroReq{
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Duration:    req.Duration,
		ArchivedAt:  req.ArchivedAt,
	}
}

// toListPomodoroInput 将获取常用番茄工作列表请求转换为应用层入参
// @param req 获取常用番茄工作列表请求
// @return 应用层获取常用番茄工作列表入参
func toListPomodoroInput(req types.ListPomodoroReq) *pomodoroDto.ListPomodoroReq {
	return &pomodoroDto.ListPomodoroReq{
		Type:       req.Type,
		Name:       req.Name,
		IsArchived: req.IsArchived,
		Page:       req.Page,
		Limit:      req.Limit,
		Sort:       req.Sort,
	}
}

// toPomodoroRes 将应用层常用番茄工作出参转换为常用番茄工作响应
// @param output 应用层常用番茄工作出参
// @return 常用番茄工作响应
func toPomodoroRes(output *pomodoroDto.PomodoroRes) *types.PomodoroRes {
	return &types.PomodoroRes{
		ResBase: types.ResBase{
			Id:        output.Id,
			CreatedAt: output.CreatedAt,
			UpdatedAt: output.UpdatedAt,
			DeletedAt: output.DeletedAt,
		},
		Type:          output.Type,
		Name:          output.Name,
		Description:   output.Description,
		Duration:      output.Duration,
		ArchivedAt:    output.ArchivedAt,
		TotalDuration: output.TotalDuration,
	}
}

// toCreatePomodoroRes 将应用层创建常用番茄工作出参转换为创建常用番茄工作响应
// @param output 应用层创建常用番茄工作出参
// @return 创建常用番茄工作响应
func toCreatePomodoroRes(output *pomodoroDto.CreatePomodoroRes) *types.CreatePomodoroRes {
	res := toPomodoroRes((*pomodoroDto.PomodoroRes)(output))
	return (*types.CreatePomodoroRes)(res)
}

// toPomodoroReses 将应用层常用番茄工作列表出参转换为常用番茄工作列表响应
// @param output 应用层常用番茄工作列表出参
// @return 常用番茄工作列表响应
func toPomodoroReses(output pomodoroDto.ListPomodoroRes) types.ListPomodoroRes {
	res := make(types.ListPomodoroRes, 0, len(output))
	for i := range output {
		res = append(res, *toPomodoroRes(&output[i]))
	}
	return res
}

// CreatePomodoroRecord 创建番茄工作记录控制器
// @code 7001x
func (c *PomodoroController) CreatePomodoroRecord(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70013,
			Message: "用户未登录",
		})
		return
	}
	var req types.CreatePomodoroRecordReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70011,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	res, err := c.pomodoroApp.Create(
		ctx.Request.Context(),
		userId,
		toCreatePomodoroRecordInput(req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70012,
			Message: "创建专注记录失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70010,
		Message: "创建专注记录成功",
		Data:    toCreatePomodoroRecordRes(res),
	})
}

// GetPomodoroRecord 获取专注记录详情控制器
// @code 7002x
func (c *PomodoroController) GetPomodoroRecord(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70023,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70021,
			Message: "专注记录 ID 无效",
		})
		return
	}
	res, err := c.pomodoroApp.Get(
		ctx.Request.Context(),
		userId,
		toGetPomodoroRecordInput(types.GetPomodoroRecordReq{Id: id}),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70022,
			Message: err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70020,
		Message: "获取专注记录成功",
		Data:    toGetPomodoroRecordRes(res),
	})
}

// ListPomodoroRecord 获取专注记录列表控制器
// @code 7003x
func (c *PomodoroController) ListPomodoroRecord(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70033,
			Message: "用户未登录",
		})
		return
	}
	var req types.ListPomodoroRecordReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70031,
			Message: err.Error(),
		})
		return
	}
	res, total, err := c.pomodoroApp.List(
		ctx.Request.Context(),
		userId,
		toListPomodoroRecordInput(req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70032,
			Message: err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70030,
		Message: "获取专注记录列表成功",
		Data:    toGetPomodoroRecordReses(res),
		Pagination: &types.Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: total,
		},
	})
}

// --- Pomodoro ---

// GetPomodoro 获取常用番茄工作详情控制器
// @code 7004x
func (c *PomodoroController) GetPomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70043,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70041,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	res, err := c.pomodoroApp.GetPomodoro(
		ctx.Request.Context(),
		userId,
		toGetPomodoroInput(types.GetPomodoroReq{Id: id}),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70042,
			Message: err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70040,
		Message: "获取常用番茄工作成功",
		Data:    toPomodoroRes(res),
	})
}

// CreatePomodoro 创建常用番茄工作控制器
// @code 7005x
func (c *PomodoroController) CreatePomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70053,
			Message: "用户未登录",
		})
		return
	}
	var req types.CreatePomodoroReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70051,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	res, err := c.pomodoroApp.CreatePomodoro(
		ctx.Request.Context(),
		userId,
		toCreatePomodoroInput(req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70052,
			Message: "创建常用番茄工作失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70050,
		Message: "创建常用番茄工作成功",
		Data:    toCreatePomodoroRes(res),
	})
}

// UpdatePomodoro 更新常用番茄工作控制器
// @code 7006x
func (c *PomodoroController) UpdatePomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70063,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70061,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	var req types.UpdatePomodoroReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70061,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	err = c.pomodoroApp.UpdatePomodoro(
		ctx.Request.Context(),
		userId,
		id,
		toUpdatePomodoroInput(req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70062,
			Message: "更新常用番茄工作失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70060,
		Message: "更新常用番茄工作成功",
		Data:    id,
	})
}

// DeletePomodoro 删除常用番茄工作控制器
// @code 7010x
func (c *PomodoroController) DeletePomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70103,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70101,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.DeletePomodoro(ctx.Request.Context(), userId, id)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70102,
			Message: "删除常用番茄工作失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70100,
		Message: "删除常用番茄工作成功",
		Data:    id,
	})
}

// ArchivedPomodoro 归档常用番茄工作控制器
// @code 7007x
func (c *PomodoroController) ArchivedPomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70073,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70071,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.ArchivePomodoro(ctx.Request.Context(), userId, id)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70072,
			Message: "归档常用番茄工作失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70070,
		Message: "归档常用番茄工作成功",
		Data:    id,
	})
}

// UnarchivedPomodoro 取消归档常用番茄工作控制器
// @code 7008x
func (c *PomodoroController) UnarchivedPomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70083,
			Message: "用户未登录",
		})
		return
	}
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70081,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.UnarchivePomodoro(ctx.Request.Context(), userId, id)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70082,
			Message: "取消归档常用番茄工作失败",
			Error:   err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70080,
		Message: "取消归档常用番茄工作成功",
		Data:    id,
	})
}

// ListPomodoro 获取常用番茄工作列表控制器
// @code 7009x
func (c *PomodoroController) ListPomodoro(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    70093,
			Message: "用户未登录",
		})
		return
	}
	var req types.ListPomodoroReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70091,
			Message: err.Error(),
		})
		return
	}
	res, total, err := c.pomodoroApp.ListPomodoro(
		ctx.Request.Context(),
		userId,
		toListPomodoroInput(req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70092,
			Message: err.Error(),
		})
		return
	}
	Success(ctx, types.ResponseData{
		Code:    70090,
		Message: "获取常用番茄工作列表成功",
		Data:    toPomodoroReses(res),
		Pagination: &types.Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: total,
		},
	})
}
