package controllers

import (
	"fmt"
	"strconv"
	"time"

	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// SyncController 数据同步控制器
type SyncController struct {
	taskApp      taskApp.TaskApp
	checkItemApp taskApp.TaskCheckItemApp
	commentApp   taskApp.TaskCommentApp
	projectApp   projectApp.ProjectApp
	tagApp       tagApp.TagApp
	pomodoroApp  pomodoroApp.PomodoroApp
}

// NewSyncController 创建数据同步控制器实例
func NewSyncController(
	task taskApp.TaskApp,
	checkItem taskApp.TaskCheckItemApp,
	comment taskApp.TaskCommentApp,
	project projectApp.ProjectApp,
	tag tagApp.TagApp,
	pomodoro pomodoroApp.PomodoroApp,
) *SyncController {
	return &SyncController{
		taskApp:      task,
		checkItemApp: checkItem,
		commentApp:   comment,
		projectApp:   project,
		tagApp:       tag,
		pomodoroApp:  pomodoro,
	}
}

// Push 批量推送控制器
// @code 9001x
// 多表批量 upsert（客户端预置 id 时走 LWW 幂等）+ 删除墓碑同步
func (c *SyncController) Push(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    90013,
			Message: "用户未登录",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.SyncPushReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    90011,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 逐表幂等写入（app 层 Create 已支持客户端指定 id 的 upsert）
	results := make([]types.SyncResult, 0,
		len(req.Tasks)+len(req.TaskCheckItems)+len(req.TaskComments)+
			len(req.Projects)+len(req.Tags)+len(req.Pomodoros)+len(req.PomodoroRecords)+len(req.Deletions))

	for i := range req.Tasks {
		res, err := c.taskApp.CreateTask(ctx.Request.Context(), userId, toCreateTaskReq(&req.Tasks[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "tasks", Id: syncID(req.Tasks[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "tasks", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.TaskCheckItems {
		res, err := c.checkItemApp.CreateTaskCheckItem(ctx.Request.Context(), userId, toCreateTaskCheckItemReq(&req.TaskCheckItems[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "taskCheckItems", Id: syncID(req.TaskCheckItems[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "taskCheckItems", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.TaskComments {
		res, err := c.commentApp.CreateTaskComment(ctx.Request.Context(), userId, toCreateTaskCommentReq(&req.TaskComments[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "taskComments", Id: syncID(req.TaskComments[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "taskComments", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.Projects {
		res, err := c.projectApp.Create(ctx.Request.Context(), userId, toCreateProjectInput(&req.Projects[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "projects", Id: syncID(req.Projects[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "projects", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.Tags {
		res, err := c.tagApp.CreateTag(ctx.Request.Context(), userId, toCreateTagInput(&req.Tags[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "tags", Id: syncID(req.Tags[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "tags", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.Pomodoros {
		res, err := c.pomodoroApp.CreatePomodoro(ctx.Request.Context(), userId, toCreatePomodoroInput(req.Pomodoros[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "pomodoros", Id: syncID(req.Pomodoros[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "pomodoros", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}
	for i := range req.PomodoroRecords {
		res, err := c.pomodoroApp.Create(ctx.Request.Context(), userId, toCreatePomodoroRecordInput(req.PomodoroRecords[i]))
		if err != nil {
			results = append(results, types.SyncResult{Table: "pomodoroRecords", Id: syncID(req.PomodoroRecords[i].Id), Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: "pomodoroRecords", Id: res.Id, ServerUpdatedAt: res.UpdatedAt})
	}

	// 4. 删除墓碑（软删，服务端推进 updated_at）
	now := time.Now()
	for _, d := range req.Deletions {
		var err error
		switch d.Table {
		case "tasks":
			err = c.taskApp.DeleteTask(ctx.Request.Context(), userId, d.Id)
		case "taskCheckItems":
			err = c.checkItemApp.DeleteTaskCheckItem(ctx.Request.Context(), userId, d.Id)
		case "taskComments":
			err = c.commentApp.DeleteTaskComment(ctx.Request.Context(), userId, d.Id)
		case "projects":
			err = c.projectApp.Delete(ctx.Request.Context(), userId, d.Id)
		case "tags":
			err = c.tagApp.DeleteTag(ctx.Request.Context(), userId, d.Id)
		case "pomodoros":
			err = c.pomodoroApp.DeletePomodoro(ctx.Request.Context(), userId, d.Id)
		case "pomodoroRecords":
			// PomodoroRecord 为只追加记录，无删除接口，忽略删除请求并明确标记
			results = append(results, types.SyncResult{Table: d.Table, Id: d.Id, Skipped: true})
			continue
		default:
			err = fmt.Errorf("未知删除表: %s", d.Table)
		}
		if err != nil {
			results = append(results, types.SyncResult{Table: d.Table, Id: d.Id, Error: err.Error()})
			continue
		}
		results = append(results, types.SyncResult{Table: d.Table, Id: d.Id, ServerUpdatedAt: now.Format(time.RFC3339)})
	}

	// 5. 返回结果
	Success(ctx, types.ResponseData{
		Code:    90010,
		Message: "批量推送成功",
		Data: types.SyncPushRes{
			Results:    results,
			ServerTime: strconv.FormatInt(time.Now().UnixMilli(), 10),
		},
	})
}

// Pull 批量增量拉取控制器
// @code 9002x
// 按表 updatedAt 游标拉取（含软删墓碑、稳定排序），响应 nextCursor 与 serverTime
func (c *SyncController) Pull(ctx *gin.Context) {
	// 1. 获取当前登录用户 ID
	userId := iCtx.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    90023,
			Message: "用户未登录",
		})
		return
	}
	// 2. 绑定请求参数
	var req types.SyncPullReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		Failure(ctx, types.ResponseData{
			Code:    90021,
			Message: "参数错误",
			Error:   err.Error(),
		})
		return
	}
	// 3. 按表增量拉取
	data := make(map[string]types.SyncPullTableRes)

	if t := req.Tasks; t != nil {
		items, err := c.taskApp.ListTaskSync(ctx.Request.Context(), userId, &taskDto.ListTaskReq{
			UpdatedAt: t.UpdatedAt, CursorId: t.CursorId, Limit: t.Limit,
		})
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "任务增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetTaskResList(items)
		data["tasks"] = types.SyncPullTableRes{Items: res, Total: int64(len(res)), NextCursor: lastUpdatedAt(res), NextCursorId: lastIdOf(res)}
	}
	if t := req.Projects; t != nil {
		items, err := c.projectApp.ListSync(ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "清单增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetProjectResList(items)
		data["projects"] = types.SyncPullTableRes{Items: res, Total: int64(len(res)), NextCursor: lastProjectUpdatedAt(res), NextCursorId: lastProjectIdOf(res)}
	}
	if t := req.Tags; t != nil {
		items, err := c.tagApp.ListTagSync(ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "标签增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetTagResList(items)
		data["tags"] = types.SyncPullTableRes{Items: res, Total: int64(len(res)), NextCursor: lastTagUpdatedAt(res), NextCursorId: lastTagIdOf(res)}
	}
	if t := req.Pomodoros; t != nil {
		items, err := c.pomodoroApp.ListPomodoroSync(ctx.Request.Context(), userId, &pomodoroDto.ListPomodoroReq{
			UpdatedAt: t.UpdatedAt, CursorId: t.CursorId, Limit: t.Limit,
		})
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "常用番茄增量拉取失败", Error: err.Error()})
			return
		}
		res := toPomodoroReses(items)
		data["pomodoros"] = types.SyncPullTableRes{Items: res, Total: int64(len(res)), NextCursor: lastPomodoroUpdatedAt(res), NextCursorId: lastPomodoroIdOf(res)}
	}
	if t := req.PomodoroRecords; t != nil {
		items, err := c.pomodoroApp.ListSync(ctx.Request.Context(), userId, &pomodoroDto.ListPomodoroRecordReq{
			UpdatedAt: t.UpdatedAt, CursorId: t.CursorId, Limit: t.Limit,
		})
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "番茄记录增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetPomodoroRecordReses(items)
		data["pomodoroRecords"] = types.SyncPullTableRes{Items: res, Total: int64(len(res)), NextCursor: lastPomodoroRecordUpdatedAt(res), NextCursorId: lastPomodoroRecordIdOf(res)}
	}

	// 4. 返回结果
	Success(ctx, types.ResponseData{
		Code:    90020,
		Message: "批量拉取成功",
		Data: types.SyncPullRes{
			Data:       data,
			ServerTime: strconv.FormatInt(time.Now().UnixMilli(), 10),
		},
	})
}

// lastUpdatedAt 取任务列表最后一条的 UpdatedAt（RFC3339）作为下一页游标；空列表返回空
func lastUpdatedAt(resList []*types.GetTaskRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastProjectUpdatedAt 取项目列表最后一条 updatedAt
func lastProjectUpdatedAt(resList []*types.GetProjectRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastTagUpdatedAt 取标签列表最后一条 updatedAt
func lastTagUpdatedAt(resList []*types.GetTagRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastPomodoroUpdatedAt 取常用番茄列表最后一条 updatedAt
func lastPomodoroUpdatedAt(resList []types.PomodoroRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastPomodoroRecordUpdatedAt 取番茄记录列表最后一条 updatedAt
func lastPomodoroRecordUpdatedAt(resList []*types.GetPomodoroRecordRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastIdOf 取任务列表最后一条 id（keyset 游标辅助）
func lastIdOf(resList []*types.GetTaskRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// lastProjectIdOf 取项目列表最后一条 id
func lastProjectIdOf(resList []*types.GetProjectRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// lastTagIdOf 取标签列表最后一条 id
func lastTagIdOf(resList []*types.GetTagRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// lastPomodoroIdOf 取常用番茄列表最后一条 id
func lastPomodoroIdOf(resList []types.PomodoroRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// lastPomodoroRecordIdOf 取番茄记录列表最后一条 id
func lastPomodoroRecordIdOf(resList []*types.GetPomodoroRecordRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// syncID 取客户端预置 id（推送失败条目标识），未提供返回空串
func syncID(id *string) string {
	if id == nil {
		return ""
	}
	return *id
}
