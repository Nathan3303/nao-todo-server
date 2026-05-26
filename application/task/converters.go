package task

import (
	"naotodoserver/consts"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
)

// weekdaysToBitmask 星期数组转换位掩码
// @param weekdays 星期数组
// @return int8 位掩码
func weekdaysToBitmask(weekdays []int) int8 {
	var mask int8
	for _, d := range weekdays {
		if v, ok := consts.WeekdayBitmask[d]; ok {
			mask |= v
		}
	}
	return mask
}

// bitmaskToWeekdays 位掩码转换星期数组
// @param mask 位掩码
// @return []int 星期数组
func bitmaskToWeekdays(mask int8) []int {
	var weekdays []int
	for bit, day := range consts.WeekdayBitmaskReverse {
		if mask&bit != 0 {
			weekdays = append(weekdays, day)
		}
	}
	if weekdays == nil {
		weekdays = []int{}
	}
	return weekdays
}

// TaskEntityToGetRes 任务实体转换为获取任务响应
// @param taskEntity 任务实体
// @return 任务响应
func TaskEntityToGetRes(taskEntity *entities.Task) *types.GetTaskRes {
	res := &types.GetTaskRes{}
	res.Id = strconv.FormatInt(taskEntity.Id, 10)
	res.ParentTaskId = strconv.FormatInt(taskEntity.ParentTaskId, 10)
	res.Name = taskEntity.Name
	res.Description = taskEntity.Description
	res.State = consts.TodoStateMapReverse[taskEntity.State]
	res.Priority = consts.TodoPriorityMapReverse[taskEntity.Priority]
	res.StartAt = taskEntity.GetFormatedStartAt()
	res.EndAt = taskEntity.GetFormatedEndAt()
	res.ProjectId = strconv.FormatInt(taskEntity.ProjectId, 10)
	res.Tags = taskEntity.Tags
	res.ArchivedAt, _ = taskEntity.ParseArchivedAt()
	res.StarMarkAt, _ = taskEntity.ParseStarMarkAt()
	res.GivenUpAt, _ = taskEntity.ParseGivenUpAt()
	res.RemindAt = taskEntity.GetFormatedRemindAt()
	res.RemindRepeat = consts.RemindRepeatMapReverse[taskEntity.RemindRepeat]
	res.RemindTime = taskEntity.RemindTime
	res.RemindWeekdays = bitmaskToWeekdays(taskEntity.RemindWeekdays)
	res.UpdatedAt = utils.Time2String(taskEntity.UpdatedAt)
	res.CreatedAt = utils.Time2String(taskEntity.CreatedAt)
	res.DeletedAt = utils.Time2String(taskEntity.DeletedAt)
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
	parentTaskIdInt64, err := strconv.ParseInt(req.ParentTaskId, 10, 64)
	if err != nil {
		parentTaskIdInt64 = 0
	}
	var projectIdInt64 int64
	if req.ProjectId == "" {
		projectIdInt64 = userId
	} else {
		projectIdInt64, err = strconv.ParseInt(req.ProjectId, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return valueobjects.NewCreateTask(
		parentTaskIdInt64,
		req.Name,
		req.Description,
		consts.TodoStateMap[req.State],
		consts.TodoPriorityMap[req.Priority],
		utils.String2SqlNullTime(req.StartAt),
		utils.String2SqlNullTime(req.EndAt),
		projectIdInt64,
		req.Tags,
		utils.String2SqlNullTime(req.RemindAt),
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
	var iState, iPriority *int8
	var iRemindRepeat *int8
	var iRemindWeekdays *int8
	if req.ParentTaskId != nil {
		iParentIdValue, _ := strconv.ParseInt(*req.ParentTaskId, 10, 64)
		iParentId = &iParentIdValue
	}
	if req.ProjectId != nil {
		if *req.ProjectId == "inbox" {
			iProjectId = &userId
		} else {
			iProjectIdValue, _ := strconv.ParseInt(*req.ProjectId, 10, 64)
			iProjectId = &iProjectIdValue
		}
	}
	if req.State != nil {
		iStateValue := consts.TodoStateMap[*req.State]
		iState = &iStateValue
	}
	if req.Priority != nil {
		iPriorityValue := consts.TodoPriorityMap[*req.Priority]
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
		utils.NullableString2NullableTime(req.StartAt),
		utils.NullableString2NullableTime(req.EndAt),
		iProjectId,
		req.Tags,
		utils.NullableString2NullableTime(req.ArchivedAt),
		utils.NullableString2NullableTime(req.StarMarkAt),
		utils.NullableString2NullableTime(req.GivenUpAt),
		utils.NullableString2NullableTime(req.RemindAt),
		iRemindRepeat,
		req.RemindTime,
		iRemindWeekdays,
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
		porjectIdValue, err := strconv.ParseInt(req.ProjectId, 10, 64)
		if err != nil {
			projectIdInt64 = 0
		} else {
			projectIdInt64 = porjectIdValue
		}
	}
	return valueobjects.NewQueryTask(
		userId,
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
