package task

import (
	"naotodoserver/application/idutil"
	"naotodoserver/consts"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/interfaces/types"
	"time"
)

// weekdaysToBitmask 星期数组转换位掩码
// @param weekdays 星期数组
// @return uint8 位掩码
func weekdaysToBitmask(weekdays []uint8) uint8 {
	var mask uint8
	for _, d := range weekdays {
		if v, ok := consts.WeekdayBitmask[int(d)]; ok {
			mask |= v
		}
	}
	return mask
}

// bitmaskToWeekdays 位掩码转换星期数组
// @param mask 位掩码
// @return []uint8 星期数组
func bitmaskToWeekdays(mask uint8) []uint8 {
	var weekdays []uint8
	for bit, day := range consts.WeekdayBitmaskReverse {
		if mask&bit != 0 {
			weekdays = append(weekdays, day)
		}
	}
	if weekdays == nil {
		weekdays = []uint8{}
	}
	return weekdays
}

// TaskEntityToGetRes 任务实体转换为获取任务响应
func TaskEntityToGetRes(taskEntity *entities.Task) *types.GetTaskRes {
	res := &types.GetTaskRes{}
	res.Id = idutil.FormatID(taskEntity.Id)
	res.UpdatedAt = taskEntity.UpdatedAt.Format(time.RFC3339)
	res.CreatedAt = taskEntity.CreatedAt.Format(time.RFC3339)
	res.DeletedAt = taskEntity.DeletedAt.ToString(time.RFC3339)
	res.ParentTaskId = idutil.FormatID(taskEntity.ParentTaskId)
	if taskEntity.ParentTaskId == 0 {
		res.ParentTaskId = ""
	}
	res.Name = taskEntity.Name
	res.Description = taskEntity.Description
	res.State = consts.TodoStateMapReverse[uint8(taskEntity.State)]
	res.Priority = consts.TodoPriorityMapReverse[uint8(taskEntity.Priority)]
	res.StartAt = taskEntity.StartAt.ToString(time.RFC3339)
	res.EndAt = taskEntity.EndAt.ToString(time.RFC3339)
	res.ProjectId = idutil.FormatID(taskEntity.ProjectId)
	res.Tags = taskEntity.Tags
	res.ArchivedAt = taskEntity.ArchivedAt.ToString(time.RFC3339)
	res.StarMarkAt = taskEntity.StarMarkAt.ToString(time.RFC3339)
	res.GivenUpAt = taskEntity.GivenUpAt.ToString(time.RFC3339)
	res.RemindAt = taskEntity.RemindAt.ToString(time.RFC3339)
	res.RemindRepeat = consts.RemindRepeatMapReverse[taskEntity.RemindRepeat]
	res.RemindTime = taskEntity.RemindTime
	res.RemindWeekdays = bitmaskToWeekdays(taskEntity.RemindWeekdays)
	res.SortId = taskEntity.SortId
	return res
}

// CreateTaskReqToValueObject 创建任务请求转换为创建任务值对象
// @param userId 用户 ID
// @param req 创建任务请求
// @return 创建任务值对象
// @error 错误
func CreateTaskReqToValueObject(
	userId int64,
	req *types.CreateTaskReq,
) (*valueobjects.CreateTask, error) {
	parentTaskIdInt64, err := idutil.ParseID(req.ParentTaskId)
	if err != nil {
		parentTaskIdInt64 = 0
	}
	var projectIdInt64 int64
	if req.ProjectId == "" {
		projectIdInt64 = userId
	} else {
		projectIdInt64, err = idutil.ParseID(req.ProjectId)
		if err != nil {
			return nil, err
		}
	}
	return valueobjects.NewCreateTask(
		parentTaskIdInt64,
		req.Name,
		req.Description,
		entities.TaskState(consts.TodoStateMap[req.State]),
		entities.TaskPriority(consts.TodoPriorityMap[req.Priority]),
		req.StartAt,
		req.EndAt,
		projectIdInt64,
		req.Tags,
		req.RemindAt,
		consts.RemindRepeatMap[req.RemindRepeat],
		req.RemindTime,
		weekdaysToBitmask(req.RemindWeekdays),
	)
}

// UpdateTaskReqToValueObject 更新任务请求转换为更新任务值对象
// @param userId 用户 ID
// @param req 更新任务请求
// @return 更新任务值对象
// @error 错误
func UpdateTaskReqToValueObject(
	userId int64,
	req *types.UpdateTaskReq,
) (*valueobjects.UpdateTask, error) {
	var iParentId, iProjectId *int64
	var iState *entities.TaskState
	var iPriority *entities.TaskPriority
	var iRemindRepeat *uint8
	var iRemindWeekdays *uint8
	if req.ParentTaskId != nil {
		iParentIdValue, _ := idutil.ParseID(*req.ParentTaskId)
		iParentId = &iParentIdValue
	}
	if req.ProjectId != nil {
		if *req.ProjectId == "inbox" {
			iProjectId = &userId
		} else {
			iProjectIdValue, _ := idutil.ParseID(*req.ProjectId)
			iProjectId = &iProjectIdValue
		}
	}
	if req.State != nil {
		iStateValue := entities.TaskState(consts.TodoStateMap[*req.State])
		iState = &iStateValue
	}
	if req.Priority != nil {
		iPriorityValue := entities.TaskPriority(consts.TodoPriorityMap[*req.Priority])
		iPriority = &iPriorityValue
	}
	if req.RemindRepeat != nil {
		iRemindRepeatValue := consts.RemindRepeatMap[*req.RemindRepeat]
		iRemindRepeat = &iRemindRepeatValue
	}
	if req.RemindWeekdays != nil {
		iRemindWeekdaysValue := weekdaysToBitmask(req.RemindWeekdays)
		iRemindWeekdays = &iRemindWeekdaysValue
	}
	return valueobjects.NewUpdateTask(
		0,
		iParentId,
		req.Name,
		req.Description,
		iState,
		iPriority,
		req.StartAt,
		req.EndAt,
		iProjectId,
		req.Tags,
		req.ArchivedAt,
		req.StarMarkAt,
		req.GivenUpAt,
		req.RemindAt,
		iRemindRepeat,
		req.RemindTime,
		iRemindWeekdays,
		req.SortId,
	)
}

// ListTaskReqToQueryTaskValueObject 列表任务请求转换为查询任务值对象
// @param userId 用户 ID
// @param req 列表任务请求
// @return 查询任务值对象
// @error 错误
func ListTaskReqToQueryTaskValueObject(
	userId int64,
	req *types.ListTaskReq,
) (*valueobjects.QueryTask, error) {
	var projectIdInt64 int64
	if req.ProjectId == "inbox" {
		projectIdInt64 = userId
	} else {
		porjectIdValue, err := idutil.ParseID(req.ProjectId)
		if err != nil {
			projectIdInt64 = 0
		} else {
			projectIdInt64 = porjectIdValue
		}
	}
	parentTaskIdInt64, _ := idutil.ParseID(req.ParentTaskId)
	return valueobjects.NewQueryTask(
		userId,
		parentTaskIdInt64,
		projectIdInt64,
		req.TagId,
		req.Name,
		req.Description,
		req.State,
		req.Priority,
		req.StartAt,
		req.EndAt,
		req.DeletedAt,
		req.ArchivedAt,
		req.StarMarkAt,
		req.GivenUpAt,
		req.IsDeleted,
		req.IsArchived,
		req.IsStarMarked,
		req.IsGivenUp,
		req.Page,
		req.Limit,
		req.RelativeDate,
		req.Sort,
	)
}

// TaskEntitiesToGetReses 任务实体转换为获取任务响应列表
// @param taskEntities 任务实体列表
// @return 任务响应列表
func TaskEntitiesToGetReses(taskEntities []*entities.Task) []*types.GetTaskRes {
	reses := make([]*types.GetTaskRes, 0, len(taskEntities))
	for _, item := range taskEntities {
		reses = append(reses, TaskEntityToGetRes(item))
	}
	return reses
}

// PaginationValueObjectToRes 分页值对象转换为分页响应
// @param paginationValueObject 分页值对象
// @return 分页响应
func PaginationValueObjectToRes(paginationValueObject *valueobjects.Pagination) *types.Pagination {
	paginationValueObject.CalcMaxPage()
	return &types.Pagination{
		Page:    paginationValueObject.Page,
		Limit:   paginationValueObject.Limit,
		Total:   paginationValueObject.Total,
		MaxPage: paginationValueObject.MaxPage,
	}
}

// --- TaskCheckItem converters ---

// TaskCheckItemEntityToGetRes 任务检查项实体转换为获取任务检查项响应
// @param e 任务检查项实体
// @return 任务检查项响应
func TaskCheckItemEntityToGetRes(e *entities.TaskCheckItem) *types.GetTaskCheckItemRes {
	var res types.GetTaskCheckItemRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.TaskId = idutil.FormatID(e.TaskId)
	res.Name = e.Name
	res.Description = e.Description
	res.IsDone = e.IsDone
	res.SortId = e.SortId
	return &res
}

// CreateTaskCheckItemReqToVO 创建任务检查项请求转换为创建任务检查项值对象
// @param userId 用户 ID
// @param req 创建任务检查项请求
// @return 创建任务检查项值对象
// @error 错误
func CreateTaskCheckItemReqToVO(
	userId int64,
	req *types.CreateTaskCheckItemReq,
) (*valueobjects.CreateTaskCheckItem, error) {
	taskId, err := idutil.ParseID(req.TaskId)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateTaskCheckItem(
		userId,
		taskId,
		req.Name,
		req.Description,
	)
}

// TaskCheckItemEntityToCreateRes 任务检查项实体转换为创建任务检查项响应
// @param e 任务检查项实体
// @return 创建任务检查项响应
func TaskCheckItemEntityToCreateRes(e *entities.TaskCheckItem) *types.CreateTaskCheckItemRes {
	var res types.CreateTaskCheckItemRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.TaskId = idutil.FormatID(e.TaskId)
	res.Name = e.Name
	res.Description = e.Description
	res.IsDone = e.IsDone
	res.SortId = e.SortId
	return &res
}

// UpdateTaskCheckItemReqToVO 更新任务检查项请求转换为更新任务检查项值对象
// @param req 更新任务检查项请求
// @return 更新任务检查项值对象
// @error 错误
func UpdateTaskCheckItemReqToVO(
	req *types.UpdateTaskCheckItemReq,
) (*valueobjects.UpdateTaskCheckItem, error) {
	return valueobjects.NewUpdateTaskCheckItem(
		req.Name,
		req.Description,
		req.IsDone,
		req.SortId,
	)
}

// TaskCheckItemEntitiesToReses 任务检查项实体转换为获取任务检查项响应列表
// @param items 任务检查项实体列表
// @return 任务检查项响应列表
func TaskCheckItemEntitiesToReses(items []*entities.TaskCheckItem) types.ListTaskCheckItemRes {
	res := make([]*types.GetTaskCheckItemRes, 0, len(items))
	for _, e := range items {
		res = append(res, TaskCheckItemEntityToGetRes(e))
	}
	return res
}

// BatchUpdateTaskCheckItemReqToVOs 批量更新任务检查项请求转换为批量更新任务检查项值对象列表
// @param req 批量更新任务检查项请求
// @return 批量更新任务检查项值对象列表
// @error 错误
func BatchUpdateTaskCheckItemReqToVOs(
	req *types.BatchUpdateTaskCheckItemReq,
) ([]*valueobjects.BatchUpdateTaskCheckItem, error) {
	vos := make([]*valueobjects.BatchUpdateTaskCheckItem, 0, len(req.Events))
	for _, e := range req.Events {
		id, err := idutil.ParseID(e.Id)
		if err != nil {
			return nil, err
		}
		vo, err := valueobjects.NewBatchUpdateTaskCheckItem(
			id,
			e.Name,
			e.Description,
			e.IsDone,
			e.SortId,
		)
		if err != nil {
			return nil, err
		}
		vos = append(vos, vo)
	}
	return vos, nil
}

// --- TaskComment converters ---

// TaskCommentEntityToRes 任务评论实体转换为获取任务评论响应
// @param e 任务评论实体
// @return 任务评论响应
func TaskCommentEntityToRes(e *entities.TaskComment) *types.TaskCommentRes {
	var res types.TaskCommentRes
	res.Id = idutil.FormatID(e.Id)
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = e.DeletedAt.ToString(time.RFC3339)
	res.TaskId = idutil.FormatID(e.TaskId)
	res.Content = e.Content
	res.Attachments = e.Attachments
	res.IsTopUp = e.IsTopUp
	res.Nickname = e.Nickname
	res.Avatar = e.Avatar
	return &res
}

// CreateTaskCommentReqToVO 创建任务评论请求转换为创建任务评论值对象
// @param userId 用户 ID
// @param req 创建任务评论请求
// @return 创建任务评论值对象
// @error 错误
func CreateTaskCommentReqToVO(
	userId int64,
	req *types.CreateTaskCommentReq,
) (*valueobjects.CreateTaskComment, error) {
	taskId, err := idutil.ParseID(req.TaskId)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateTaskComment(
		userId,
		taskId,
		req.Content,
		nil,
		false,
	)
}

// UpdateTaskCommentReqToVO 更新任务评论请求转换为更新任务评论值对象
// @param req 更新任务评论请求
// @return 更新任务评论值对象
// @error 错误
func UpdateTaskCommentReqToVO(
	req *types.UpdateTaskCommentReq,
) (*valueobjects.UpdateTaskComment, error) {
	return valueobjects.NewUpdateTaskComment(
		req.Content,
		req.Attachments,
		req.IsTopUp,
	)
}

// TaskCommentEntitiesToListRes 任务评论实体列表转换为获取任务评论响应列表
// @param list 任务评论实体列表
// @return 任务评论响应列表
func TaskCommentEntitiesToListRes(list []*entities.TaskComment) []*types.TaskCommentRes {
	res := make([]*types.TaskCommentRes, 0, len(list))
	for _, e := range list {
		res = append(res, TaskCommentEntityToRes(e))
	}
	return res
}
