package task

import (
	"naotodoserver/consts"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/utils"
	"naotodoserver/interfaces/types"
	"strconv"
)

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
	return valueobjects.NewUpdateTask(
		0,
		iParentId,
		req.Name,
		req.Description,
		iState,
		iPriority,
		utils.StringPtr2SqlNullTime(req.StartAt),
		utils.StringPtr2SqlNullTime(req.EndAt),
		iProjectId,
		req.Tags,
		utils.StringPtr2SqlNullTime(req.ArchivedAt),
		utils.StringPtr2SqlNullTime(req.StarMarkAt),
		utils.StringPtr2SqlNullTime(req.GivenUpAt),
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
