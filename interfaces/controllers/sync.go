package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"naotodoserver/application/idutil"
	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domerr "naotodoserver/domain/errors"
	domaintypes "naotodoserver/domain/types"
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

// syncOutcomeOf 将领域 upsert 判定映射为同步回执语义（唯一映射点，避免两套语义）
func syncOutcomeOf(result domaintypes.UpsertResult) string {
	switch result.Outcome {
	case domaintypes.UpsertNoop:
		return types.SyncOutcomeNoop
	case domaintypes.UpsertStale:
		return types.SyncOutcomeStale
	default:
		return types.SyncOutcomeApplied
	}
}

// basePtr 返回 sync 条目 baseUpdatedAt 的指针；空串返回 nil（= 未提供，回退 LWW）
func basePtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// syncErrOutcome 失败条目语义：create 语义 ID 碰撞 = conflict，其余 = error
func syncErrOutcome(err error) string {
	if errors.Is(err, domerr.ErrIDConflict) {
		return types.SyncOutcomeConflict
	}
	return types.SyncOutcomeError
}

// Push 批量推送控制器
// @code 9001x
// 多表批量 upsert（客户端预置 id 时走 LWW 幂等）+ 删除墓碑同步
//
// 每条结果额外回传 outcome（additive，2026-09-24 T143），语义与 domain/types.DecideUpsert 判定同源：
// applied / noop / conflict / skipped / error。
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
	total := len(req.Tasks) + len(req.TaskCheckItems) + len(req.TaskComments) +
		len(req.Projects) + len(req.Tags) + len(req.Pomodoros) +
		len(req.PomodoroRecords) + len(req.Deletions)
	results := make([]types.SyncResult, 0, total)

	for i := range req.Tasks {
		taskReq := toCreateTaskReqFromSync(&req.Tasks[i])
		taskReq.BaseUpdatedAt = basePtr(req.Tasks[i].BaseUpdatedAt)
		res, upsert, err := c.taskApp.CreateTask(ctx.Request.Context(), userId, taskReq)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "tasks", Id: syncID(req.Tasks[i].Id), Error: err.Error(),
				Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "tasks", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.TaskCheckItems {
		itemReq := toCreateTaskCheckItemReq(&req.TaskCheckItems[i].CreateTaskCheckItemReq)
		itemReq.BaseUpdatedAt = basePtr(req.TaskCheckItems[i].BaseUpdatedAt)
		res, upsert, err := c.checkItemApp.CreateTaskCheckItem(
			ctx.Request.Context(), userId, itemReq,
		)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "taskCheckItems", Id: syncID(req.TaskCheckItems[i].Id),
				Error: err.Error(), Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "taskCheckItems", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.TaskComments {
		commentReq := toCreateTaskCommentReq(&req.TaskComments[i].CreateTaskCommentReq)
		commentReq.BaseUpdatedAt = basePtr(req.TaskComments[i].BaseUpdatedAt)
		res, upsert, err := c.commentApp.CreateTaskComment(
			ctx.Request.Context(), userId, commentReq,
		)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "taskComments", Id: syncID(req.TaskComments[i].Id), Error: err.Error(),
				Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "taskComments", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.Projects {
		projectReq := toCreateProjectInput(&req.Projects[i].CreateProjectReq)
		projectReq.BaseUpdatedAt = basePtr(req.Projects[i].BaseUpdatedAt)
		res, upsert, err := c.projectApp.Create(ctx.Request.Context(), userId, projectReq)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "projects", Id: syncID(req.Projects[i].Id), Error: err.Error(),
				Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "projects", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.Tags {
		tagReq := toCreateTagInput(&req.Tags[i].CreateTagReq)
		tagReq.BaseUpdatedAt = basePtr(req.Tags[i].BaseUpdatedAt)
		res, upsert, err := c.tagApp.CreateTag(ctx.Request.Context(), userId, tagReq)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "tags", Id: syncID(req.Tags[i].Id), Error: err.Error(),
				Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "tags", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.Pomodoros {
		pomodoroReq := toCreatePomodoroInput(req.Pomodoros[i].CreatePomodoroReq)
		pomodoroReq.BaseUpdatedAt = basePtr(req.Pomodoros[i].BaseUpdatedAt)
		res, upsert, err := c.pomodoroApp.CreatePomodoro(ctx.Request.Context(), userId, pomodoroReq)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "pomodoros", Id: syncID(req.Pomodoros[i].Id), Error: err.Error(),
				Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "pomodoros", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
	}
	for i := range req.PomodoroRecords {
		recordReq := toCreatePomodoroRecordInput(req.PomodoroRecords[i].CreatePomodoroRecordReq)
		recordReq.BaseUpdatedAt = basePtr(req.PomodoroRecords[i].BaseUpdatedAt)
		res, upsert, err := c.pomodoroApp.Create(ctx.Request.Context(), userId, recordReq)
		if err != nil {
			results = append(results, types.SyncResult{
				Table: "pomodoroRecords", Id: syncID(req.PomodoroRecords[i].Id),
				Error: err.Error(), Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: "pomodoroRecords", Id: res.Id, ServerUpdatedAt: res.UpdatedAt,
			Outcome: syncOutcomeOf(upsert),
		})
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
			results = append(results, types.SyncResult{
				Table: d.Table, Id: d.Id, Skipped: true, Outcome: types.SyncOutcomeSkipped,
			})
			continue
		default:
			err = fmt.Errorf("未知删除表: %s", d.Table)
		}
		if err != nil {
			results = append(results, types.SyncResult{
				Table: d.Table, Id: d.Id, Error: err.Error(), Outcome: syncErrOutcome(err),
			})
			continue
		}
		results = append(results, types.SyncResult{
			Table: d.Table, Id: d.Id, ServerUpdatedAt: idutil.FormatTimeMilli(now),
			Outcome: types.SyncOutcomeApplied,
		})
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

// syncNullableTime 将 sync 三态字段归一为共享 create DTO 的 *string 契约（复用既有清空链路）：
//   - null（显式置空）/ "" ⇒ ptr("")：应用层 nullableTimeFromCreateReq 映射为「显式清空写 NULL」
//   - 值                    ⇒ ptr(值)
//   - absent（键缺省）      ⇒ 回退内嵌 *string（JSON 绑定下必为 nil = 不写列；Go 侧程序化构造
//     保持既有语义，避免第二套清空语义）
//
// @param n 三态字段
// @param embedded 内嵌 CreateTaskReq 的同名字段（*string 版）
// @return 应用层入参使用的 *string（nil = 缺省不写列）
func syncNullableTime(n types.NullableString, embedded *string) *string {
	if !n.Present {
		return embedded
	}
	if n.Null || n.Value == "" {
		empty := ""
		return &empty
	}
	value := n.Value
	return &value
}

// toCreateTaskReqFromSync 将 /sync/push 任务条目转换为应用层入参：
// 先按共享 create 语义整体转换（非三态字段零差异），再用三态覆盖六个可空时间字段
// （startAt/endAt/archivedAt/starMarkAt/givenUpAt/remindAt，与
// valueobjects.nullableTimeFromCreateReq 的使用面同集合）。
//
// @param item sync 推送条目
// @return 应用层创建任务入参
func toCreateTaskReqFromSync(item *types.SyncTaskPushItem) *taskDto.CreateTaskReq {
	req := toCreateTaskReq(&item.CreateTaskReq)
	req.StartAt = syncNullableTime(item.StartAt, req.StartAt)
	req.EndAt = syncNullableTime(item.EndAt, req.EndAt)
	req.ArchivedAt = syncNullableTime(item.ArchivedAt, req.ArchivedAt)
	req.StarMarkAt = syncNullableTime(item.StarMarkAt, req.StarMarkAt)
	req.GivenUpAt = syncNullableTime(item.GivenUpAt, req.GivenUpAt)
	req.RemindAt = syncNullableTime(item.RemindAt, req.RemindAt)
	return req
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
		data["tasks"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastUpdatedAt(res),
			NextCursorId: lastIdOf(res),
		}
	}
	if t := req.TaskCheckItems; t != nil {
		items, err := c.checkItemApp.ListTaskCheckItemSync(
			ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit,
		)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "检查项增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetTaskCheckItemResList(items)
		data["taskCheckItems"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastCheckItemUpdatedAt(res),
			NextCursorId: lastCheckItemIdOf(res),
		}
	}
	if t := req.TaskComments; t != nil {
		items, err := c.commentApp.ListTaskCommentSync(
			ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit,
		)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "评论增量拉取失败", Error: err.Error()})
			return
		}
		res := toTaskCommentResList(items)
		data["taskComments"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastCommentUpdatedAt(res),
			NextCursorId: lastCommentIdOf(res),
		}
	}
	if t := req.Projects; t != nil {
		items, err := c.projectApp.ListSync(
			ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit,
		)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "清单增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetProjectResList(items)
		data["projects"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastProjectUpdatedAt(res),
			NextCursorId: lastProjectIdOf(res),
		}
	}
	if t := req.Tags; t != nil {
		items, err := c.tagApp.ListTagSync(
			ctx.Request.Context(), userId, t.UpdatedAt, t.CursorId, t.Limit,
		)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "标签增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetTagResList(items)
		data["tags"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastTagUpdatedAt(res),
			NextCursorId: lastTagIdOf(res),
		}
	}
	if t := req.Pomodoros; t != nil {
		listReq := pomodoroDto.ListPomodoroReq{
			UpdatedAt: t.UpdatedAt, CursorId: t.CursorId, Limit: t.Limit,
		}
		items, err := c.pomodoroApp.ListPomodoroSync(ctx.Request.Context(), userId, &listReq)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "常用番茄增量拉取失败", Error: err.Error()})
			return
		}
		res := toPomodoroReses(items)
		data["pomodoros"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastPomodoroUpdatedAt(res),
			NextCursorId: lastPomodoroIdOf(res),
		}
	}
	if t := req.PomodoroRecords; t != nil {
		listReq := pomodoroDto.ListPomodoroRecordReq{
			UpdatedAt: t.UpdatedAt, CursorId: t.CursorId, Limit: t.Limit,
		}
		items, err := c.pomodoroApp.ListSync(ctx.Request.Context(), userId, &listReq)
		if err != nil {
			Failure(ctx, types.ResponseData{Code: 90022, Message: "番茄记录增量拉取失败", Error: err.Error()})
			return
		}
		res := toGetPomodoroRecordReses(items)
		data["pomodoroRecords"] = types.SyncPullTableRes{
			Items: res, Total: int64(len(res)), NextCursor: lastPomodoroRecordUpdatedAt(res),
			NextCursorId: lastPomodoroRecordIdOf(res),
		}
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

// lastCheckItemUpdatedAt 取检查项列表最后一条 updatedAt
func lastCheckItemUpdatedAt(resList []*types.GetTaskCheckItemRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastCommentUpdatedAt 取评论列表最后一条 updatedAt
func lastCommentUpdatedAt(resList []*types.TaskCommentRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].UpdatedAt
}

// lastCheckItemIdOf 取检查项列表最后一条 id（keyset 游标辅助）
func lastCheckItemIdOf(resList []*types.GetTaskCheckItemRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
}

// lastCommentIdOf 取评论列表最后一条 id（keyset 游标辅助）
func lastCommentIdOf(resList []*types.TaskCommentRes) string {
	if len(resList) == 0 {
		return ""
	}
	return resList[len(resList)-1].Id
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
