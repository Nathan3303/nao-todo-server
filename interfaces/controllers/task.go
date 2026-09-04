package controllers

import (
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskApp taskApp.TaskApp
}

func NewTaskController(app taskApp.TaskApp) *TaskController {
	return &TaskController{taskApp: app}
}

// toGetTaskRes 将应用层获取任务出参转换为获取任务响应
// @param output 应用层获取任务出参
// @return 获取任务响应
func toGetTaskRes(output *taskDto.GetTaskRes) *types.GetTaskRes {
	res := &types.GetTaskRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.ParentTaskId = output.ParentTaskId
	res.Name = output.Name
	res.Description = output.Description
	res.State = output.State
	res.Priority = output.Priority
	res.StartAt = output.StartAt
	res.EndAt = output.EndAt
	res.Tags = output.Tags
	res.ProjectId = output.ProjectId
	res.ArchivedAt = output.ArchivedAt
	res.StarMarkAt = output.StarMarkAt
	res.GivenUpAt = output.GivenUpAt
	res.RemindAt = output.RemindAt
	res.RemindRepeat = output.RemindRepeat
	res.RemindTime = output.RemindTime
	res.RemindWeekdays = toIntWeekdays(output.RemindWeekdays)
	res.SortId = output.SortId
	return res
}

// toIntWeekdays 将内部星期数组（[]uint8）转换为接口层类型（[]int）
// 内部表示与序列化表示分离：[]uint8 即 []byte，JSON 序列化会被编码为 base64 而非数字数组
func toIntWeekdays(weekdays []uint8) []int {
	if weekdays == nil {
		return nil
	}
	res := make([]int, len(weekdays))
	for i, d := range weekdays {
		res[i] = int(d)
	}
	return res
}

// toUint8Weekdays 将接口层星期数组（[]int）转换回内部类型（[]uint8）
func toUint8Weekdays(weekdays []int) []uint8 {
	if weekdays == nil {
		return nil
	}
	res := make([]uint8, len(weekdays))
	for i, d := range weekdays {
		res[i] = uint8(d)
	}
	return res
}

// toGetTaskResList 将应用层任务列表出参转换为任务列表响应
// @param outputList 应用层任务列表出参
// @return 任务列表响应
func toGetTaskResList(outputList []*taskDto.GetTaskRes) []*types.GetTaskRes {
	resList := make([]*types.GetTaskRes, 0, len(outputList))
	for _, output := range outputList {
		resList = append(resList, toGetTaskRes(output))
	}
	return resList
}

// toCreateTaskReq 将创建任务请求转换为应用层入参
// @param req 创建任务请求
// @return 应用层创建任务入参
func toCreateTaskReq(req *types.CreateTaskReq) *taskDto.CreateTaskReq {
	return &taskDto.CreateTaskReq{
		ParentTaskId:   req.ParentTaskId,
		Name:           req.Name,
		Description:    req.Description,
		State:          req.State,
		Priority:       req.Priority,
		StartAt:        req.StartAt,
		EndAt:          req.EndAt,
		ProjectId:      req.ProjectId,
		Tags:           req.Tags,
		RemindAt:       req.RemindAt,
		RemindRepeat:   req.RemindRepeat,
		RemindTime:     req.RemindTime,
		RemindWeekdays: toUint8Weekdays(req.RemindWeekdays),
		Id:             req.Id,
		CreatedAt:      req.CreatedAt,
		UpdatedAt:      req.UpdatedAt,
		DeletedAt:      req.DeletedAt,
	}
}

// toUpdateTaskReq 将更新任务请求转换为应用层入参
// @param req 更新任务请求
// @return 应用层更新任务入参
func toUpdateTaskReq(req *types.UpdateTaskReq) *taskDto.UpdateTaskReq {
	return &taskDto.UpdateTaskReq{
		ParentTaskId:   req.ParentTaskId,
		Name:           req.Name,
		Description:    req.Description,
		State:          req.State,
		Priority:       req.Priority,
		StartAt:        req.StartAt,
		EndAt:          req.EndAt,
		ProjectId:      req.ProjectId,
		Tags:           req.Tags,
		ArchivedAt:     req.ArchivedAt,
		StarMarkAt:     req.StarMarkAt,
		GivenUpAt:      req.GivenUpAt,
		RemindAt:       req.RemindAt,
		RemindRepeat:   req.RemindRepeat,
		RemindTime:     req.RemindTime,
		RemindWeekdays: toUint8Weekdays(req.RemindWeekdays),
		SortId:         req.SortId,
		UpdatedAt:      req.UpdatedAt,
	}
}

// toListTaskReq 将列表任务请求转换为应用层入参
// @param req 列表任务请求
// @return 应用层列表任务入参
func toListTaskReq(req *types.ListTaskReq) *taskDto.ListTaskReq {
	return &taskDto.ListTaskReq{
		ParentTaskId: req.ParentTaskId,
		ProjectId:    req.ProjectId,
		TagId:        req.TagId,
		Name:         req.Name,
		Description:  req.Description,
		State:        req.State,
		Priority:     req.Priority,
		StartAt:      req.StartAt,
		EndAt:        req.EndAt,
		DeletedAt:    req.DeletedAt,
		ArchivedAt:   req.ArchivedAt,
		StarMarkAt:   req.StarMarkAt,
		GivenUpAt:    req.GivenUpAt,
		IsDeleted:    req.IsDeleted,
		IsArchived:   req.IsArchived,
		IsStarMarked: req.IsStarMarked,
		IsGivenUp:    req.IsGivenUp,
		RelativeDate: req.RelativeDate,
		Page:         req.Page,
		Limit:        req.Limit,
		Sort:         req.Sort,
		UpdatedAt:    req.UpdatedAt,
		CursorId:     req.CursorId,
	}
}

// toPagination 将应用层分页出参转换为分页响应
// @param output 应用层分页出参
// @return 分页响应
func toPagination(output *taskDto.Pagination) *types.Pagination {
	if output == nil {
		return nil
	}
	return &types.Pagination{
		Total:   output.Total,
		Page:    output.Page,
		Limit:   output.Limit,
		MaxPage: output.MaxPage,
	}
}

// toSnoozeTaskReq 将稍后提醒请求转换为应用层入参
// @param req 稍后提醒请求
// @return 应用层稍后提醒入参
func toSnoozeTaskReq(req *types.SnoozeTaskReq) *taskDto.SnoozeTaskReq {
	return &taskDto.SnoozeTaskReq{
		DurationMinutes: req.DurationMinutes,
	}
}

// toSnoozeTaskRes 将应用层稍后提醒出参转换为稍后提醒响应
// @param output 应用层稍后提醒出参
// @return 稍后提醒响应
func toSnoozeTaskRes(output *taskDto.SnoozeTaskRes) *types.SnoozeTaskRes {
	if output == nil {
		return nil
	}
	return &types.SnoozeTaskRes{
		RemindAt: output.RemindAt,
	}
}

// toGetTaskCheckItemRes 将应用层获取任务检查项出参转换为获取任务检查项响应
// @param output 应用层获取任务检查项出参
// @return 获取任务检查项响应
func toGetTaskCheckItemRes(output *taskDto.GetTaskCheckItemRes) *types.GetTaskCheckItemRes {
	res := &types.GetTaskCheckItemRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.TaskId = output.TaskId
	res.Name = output.Name
	res.Description = output.Description
	res.IsDone = output.IsDone
	res.SortId = output.SortId
	return res
}

// toGetTaskCheckItemResList 将应用层任务检查项列表出参转换为任务检查项列表响应
// @param outputList 应用层任务检查项列表出参
// @return 任务检查项列表响应
func toGetTaskCheckItemResList(
	outputList []*taskDto.GetTaskCheckItemRes,
) []*types.GetTaskCheckItemRes {
	resList := make([]*types.GetTaskCheckItemRes, 0, len(outputList))
	for _, output := range outputList {
		resList = append(resList, toGetTaskCheckItemRes(output))
	}
	return resList
}

// toCreateTaskCheckItemReq 将创建任务检查项请求转换为应用层入参
// @param req 创建任务检查项请求
// @return 应用层创建任务检查项入参
func toCreateTaskCheckItemReq(req *types.CreateTaskCheckItemReq) *taskDto.CreateTaskCheckItemReq {
	return &taskDto.CreateTaskCheckItemReq{
		TaskId:      req.TaskId,
		Name:        req.Name,
		Description: req.Description,
		Id:          req.Id,
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   req.UpdatedAt,
		IsDone:      req.IsDone,
		SortId:      req.SortId,
	}
}

// toCreateTaskCheckItemRes 将应用层创建任务检查项出参转换为创建任务检查项响应
// @param output 应用层创建任务检查项出参
// @return 创建任务检查项响应
func toCreateTaskCheckItemRes(
	output *taskDto.CreateTaskCheckItemRes,
) *types.CreateTaskCheckItemRes {
	res := &types.CreateTaskCheckItemRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.TaskId = output.TaskId
	res.Name = output.Name
	res.Description = output.Description
	res.IsDone = output.IsDone
	res.SortId = output.SortId
	return res
}

// toUpdateTaskCheckItemReq 将更新任务检查项请求转换为应用层入参
// @param req 更新任务检查项请求
// @return 应用层更新任务检查项入参
func toUpdateTaskCheckItemReq(req *types.UpdateTaskCheckItemReq) *taskDto.UpdateTaskCheckItemReq {
	return &taskDto.UpdateTaskCheckItemReq{
		Name:        req.Name,
		Description: req.Description,
		IsDone:      req.IsDone,
		SortId:      req.SortId,
		UpdatedAt:   req.UpdatedAt,
	}
}

// toBatchUpdateTaskCheckItemReq 将批量更新任务检查项请求转换为应用层入参
// @param req 批量更新任务检查项请求
// @return 应用层批量更新任务检查项入参
func toBatchUpdateTaskCheckItemReq(
	req *types.BatchUpdateTaskCheckItemReq,
) *taskDto.BatchUpdateTaskCheckItemReq {
	events := make([]*taskDto.BatchUpdateTaskCheckItemEvent, 0, len(req.Events))
	for _, event := range req.Events {
		events = append(events, &taskDto.BatchUpdateTaskCheckItemEvent{
			Id:          event.Id,
			Name:        event.Name,
			Description: event.Description,
			IsDone:      event.IsDone,
			SortId:      event.SortId,
		})
	}
	return &taskDto.BatchUpdateTaskCheckItemReq{Events: events}
}

// toBatchUpdateTaskCheckItemRes 将应用层批量更新任务检查项出参转换为批量更新任务检查项响应
// @param output 应用层批量更新任务检查项出参
// @return 批量更新任务检查项响应
func toBatchUpdateTaskCheckItemRes(
	output *taskDto.BatchUpdateTaskCheckItemRes,
) *types.BatchUpdateTaskCheckItemRes {
	return &types.BatchUpdateTaskCheckItemRes{
		UpdatedCount: output.UpdatedCount,
		Events:       toGetTaskCheckItemResList(output.Events),
	}
}

// toTaskCommentRes 将应用层任务评论出参转换为任务评论响应
// @param output 应用层任务评论出参
// @return 任务评论响应
func toTaskCommentRes(output *taskDto.TaskCommentRes) *types.TaskCommentRes {
	res := &types.TaskCommentRes{}
	res.Id = output.Id
	res.CreatedAt = output.CreatedAt
	res.UpdatedAt = output.UpdatedAt
	res.DeletedAt = output.DeletedAt
	res.TaskId = output.TaskId
	res.Content = output.Content
	res.Attachments = output.Attachments
	res.IsTopUp = output.IsTopUp
	res.Nickname = output.Nickname
	res.Avatar = output.Avatar
	return res
}

// toTaskCommentResList 将应用层任务评论列表出参转换为任务评论列表响应
// @param outputList 应用层任务评论列表出参
// @return 任务评论列表响应
func toTaskCommentResList(outputList []*taskDto.TaskCommentRes) []*types.TaskCommentRes {
	resList := make([]*types.TaskCommentRes, 0, len(outputList))
	for _, output := range outputList {
		resList = append(resList, toTaskCommentRes(output))
	}
	return resList
}

// toCreateTaskCommentReq 将创建任务评论请求转换为应用层入参
// @param req 创建任务评论请求
// @return 应用层创建任务评论入参
func toCreateTaskCommentReq(req *types.CreateTaskCommentReq) *taskDto.CreateTaskCommentReq {
	return &taskDto.CreateTaskCommentReq{
		TaskId:    req.TaskId,
		Content:   req.Content,
		Id:        req.Id,
		CreatedAt: req.CreatedAt,
		UpdatedAt: req.UpdatedAt,
	}
}

// toUpdateTaskCommentReq 将更新任务评论请求转换为应用层入参
// @param req 更新任务评论请求
// @return 应用层更新任务评论入参
func toUpdateTaskCommentReq(req *types.UpdateTaskCommentReq) *taskDto.UpdateTaskCommentReq {
	return &taskDto.UpdateTaskCommentReq{
		Content:     req.Content,
		Attachments: req.Attachments,
		IsTopUp:     req.IsTopUp,
		UpdatedAt:   req.UpdatedAt,
	}
}

// GetTask 获取待办任务详情控制器
// @code 4000x
func (c *TaskController) GetTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40003,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取待办任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40001,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 调用应用层获取任务信息
	includeDeleted := ctx.Query("isDeleted") == "true"
	res, err := c.taskApp.GetTaskById(
		ctx.Request.Context(),
		userId,
		taskId,
		includeDeleted,
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40002,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40000,
		Message: "获取待办任务详细成功",
		Data:    toGetTaskRes(res),
	})
}

// CreateTask 创建待办任务控制器
// @code 4001x
func (c *TaskController) CreateTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40013,
			Message: "用户未登录",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.CreateTaskReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40011,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 调用应用层创建任务
	res, err := c.taskApp.CreateTask(ctx.Request.Context(), userId, toCreateTaskReq(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40012,
			Message: "创建待办任务失败",
			Error:   err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40010,
		Message: "创建待办任务成功",
		Data:    toGetTaskRes(res),
	})
}

// UpdateTask 更新待办任务控制器
// @code 4002x
func (c *TaskController) UpdateTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40024,
			Message: "用户未登录",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.UpdateTaskReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40021,
			Message: err.Error(),
		})
		return
	}
	// 3. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40022,
			Message: "任务 ID 无效",
		})
		return
	}
	// 4. 调用应用层更新任务
	err = c.taskApp.UpdateTask(ctx.Request.Context(), userId, taskId, toUpdateTaskReq(&req))
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40023,
			Message: err.Error(),
		})
		return
	}
	// 5. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40020,
		Message: "更新待办任务成功",
		Data:    taskId,
	})
}

// DeleteTask 删除待办任务控制器
// @code 4003x
func (c *TaskController) DeleteTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40033,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40031,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 调用应用层删除任务
	err := c.taskApp.DeleteTask(ctx.Request.Context(), userId, taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40032,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40030,
		Message: "删除待办任务成功",
		Data:    taskId,
	})
}

// RestoreTask 恢复待办任务控制器
// @code 4004x
func (c *TaskController) RestoreTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40043,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40041,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 调用应用层恢复任务
	err := c.taskApp.RestoreTask(ctx.Request.Context(), userId, taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40042,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40040,
		Message: "恢复待办任务成功",
		Data:    taskId,
	})
}

// CopyTask 复制待办任务控制器
// @code 4006x
func (c *TaskController) CopyTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40063,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40061,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 调用应用层复制任务
	res, err := c.taskApp.CopyTask(ctx.Request.Context(), userId, taskId)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40062,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40060,
		Message: "复制待办任务成功",
		Data:    toGetTaskRes(res),
	})
}

// ListTask 获取待办任务列表控制器
// @code 4005x
func (c *TaskController) ListTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40053,
			Message: "用户未登录",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.ListTaskReq
	err := ctx.ShouldBindQuery(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40051,
			Message: err.Error(),
		})
		return
	}
	// 3. 增量同步：携带 updatedAt 游标时走增量路径（含软删墓碑、稳定排序）
	if req.UpdatedAt != "" {
		tasks, err := c.taskApp.ListTaskSync(
			ctx.Request.Context(),
			userId,
			toListTaskReq(&req),
		)
		if err != nil {
			Failure(ctx, types.ResponseData{
				Code:    40052,
				Message: err.Error(),
			})
			return
		}
		Success(ctx, types.ResponseData{
			Code:    40050,
			Message: "获取待办任务列表成功",
			Data:    toGetTaskResList(tasks),
			Pagination: &types.Pagination{
				Page:  1,
				Limit: req.Limit,
				Total: int64(len(tasks)),
			},
		})
		return
	}
	// 4. 调用应用层获取任务列表
	tasks, paginationRes, err := c.taskApp.ListTask(
		ctx.Request.Context(),
		userId,
		toListTaskReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40052,
			Message: err.Error(),
		})
		return
	}
	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:       40050,
		Message:    "获取待办任务列表成功",
		Data:       toGetTaskResList(tasks),
		Pagination: toPagination(paginationRes),
	})
}

// SnoozeTask 稍后提醒控制器
// @code 4009x
func (c *TaskController) SnoozeTask(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    40094,
			Message: "用户未登录",
		})
		return
	}
	// 2. 获取任务 ID
	taskId := ctx.Param("taskId")
	if taskId == "" {
		Failure(ctx, types.ResponseData{
			Code:    40091,
			Message: "任务 ID 无效",
		})
		return
	}
	// 3. 绑定请求参数
	var req types.SnoozeTaskReq
	err := ctx.ShouldBindJSON(&req)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40092,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 4. 调用应用层设置稍后提醒
	res, err := c.taskApp.SnoozeTask(
		ctx.Request.Context(),
		userId,
		taskId,
		toSnoozeTaskReq(&req),
	)
	if err != nil {
		Failure(ctx, types.ResponseData{
			Code:    40093,
			Message: err.Error(),
		})
		return
	}
	// 5. 返回结果
	Success(ctx, types.ResponseData{
		Code:    40090,
		Message: "稍后提醒已设置",
		Data:    toSnoozeTaskRes(res),
	})
}
