package controllers

import (
	pomodoroApp "naotodoserver/application/pomodoro"
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

// CreatePomodoroRecord 创建番茄工作记录控制器
// @code 7001x
func (c *PomodoroController) CreatePomodoroRecord(ctx *gin.Context) {
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
	res, err := c.pomodoroApp.Create(ctx.Request.Context(), &req)
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
		Data:    res,
	})
}

// GetPomodoroRecord 获取专注记录详情控制器
// @code 7002x
func (c *PomodoroController) GetPomodoroRecord(ctx *gin.Context) {
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
		&types.GetPomodoroRecordReq{Id: id},
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
		Data:    res,
	})
}

// ListPomodoroRecord 获取专注记录列表控制器
// @code 7003x
func (c *PomodoroController) ListPomodoroRecord(ctx *gin.Context) {
	var req types.ListPomodoroRecordReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70031,
			Message: err.Error(),
		})
		return
	}
	res, total, err := c.pomodoroApp.List(ctx.Request.Context(), &req)
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
		Data:    res,
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
		&types.GetPomodoroReq{Id: id},
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
		Data:    res,
	})
}

// CreatePomodoro 创建常用番茄工作控制器
// @code 7005x
func (c *PomodoroController) CreatePomodoro(ctx *gin.Context) {
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
	res, err := c.pomodoroApp.CreatePomodoro(ctx.Request.Context(), &req)
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
		Data:    res,
	})
}

// UpdatePomodoro 更新常用番茄工作控制器
// @code 7006x
func (c *PomodoroController) UpdatePomodoro(ctx *gin.Context) {
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
		id,
		&req,
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
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70101,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.DeletePomodoro(ctx.Request.Context(), id)
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
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70071,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.ArchivePomodoro(ctx.Request.Context(), id)
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
	id := ctx.Param("id")
	if id == "" {
		Failure(ctx, types.ResponseData{
			Code:    70081,
			Message: "常用番茄工作 ID 无效",
		})
		return
	}
	err := c.pomodoroApp.UnarchivePomodoro(ctx.Request.Context(), id)
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
	var req types.ListPomodoroReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    70091,
			Message: err.Error(),
		})
		return
	}
	res, total, err := c.pomodoroApp.ListPomodoro(ctx.Request.Context(), &req)
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
		Data:    res,
		Pagination: &types.Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: total,
		},
	})
}
