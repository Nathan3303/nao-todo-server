package event

import (
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
)

// EventEntityToGetRes 将检查事项实体转换为获取检查事项详情响应
// @param eventEntity 检查事项实体
// @return 获取检查事项详情响应
func EventEntityToGetRes(eventEntity *entities.Event) *types.GetEventRes {
	return &types.GetEventRes{
		Id:          strconv.FormatInt(eventEntity.Id, 10),
		TaskId:      strconv.FormatInt(eventEntity.TaskId, 10),
		Name:        eventEntity.Name,
		Description: eventEntity.Description,
		IsDone:      eventEntity.IsDone,
		SortId:      eventEntity.SortId,
		CreatedAt:   eventEntity.CreatedAt,
		UpdatedAt:   eventEntity.UpdatedAt,
	}
}

// CreateEventReqToValueObject 将创建检查事项请求体转换为创建检查事项值对象
// @param userId 用户ID
// @param createEventReq 创建检查事项请求体
// @return 创建检查事项值对象
// @return error 错误信息
func CreateEventReqToValueObject(
	userId int64,
	createEventReq *types.CreateEventReq,
) (*valueobjects.CreateEvent, error) {
	taskId, err := strconv.ParseInt(createEventReq.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateEvent(
		userId,
		taskId,
		createEventReq.Name,
		createEventReq.Description,
	)
}

// EventEntityToCreateRes 将检查事项实体转换为创建检查事项响应
// @param eventEntity 检查事项实体
// @return 创建检查事项响应
func EventEntityToCreateRes(eventEntity *entities.Event) *types.CreateEventRes {
	return &types.CreateEventRes{
		Id:          strconv.FormatInt(eventEntity.Id, 10),
		TaskId:      strconv.FormatInt(eventEntity.TaskId, 10),
		Name:        eventEntity.Name,
		Description: eventEntity.Description,
		IsDone:      eventEntity.IsDone,
		SortId:      eventEntity.SortId,
		CreatedAt:   eventEntity.CreatedAt,
		UpdatedAt:   eventEntity.UpdatedAt,
	}
}

// UpdateEventReqToValueObject 将更新检查事项请求体转换为更新检查事项值对象
// @param updateEventReq 更新检查事项请求体
// @return 更新检查事项值对象
// @return error 错误信息
func UpdateEventReqToValueObject(
	updateEventReq *types.UpdateEventReq,
) (*valueobjects.UpdateEvent, error) {
	return valueobjects.NewUpdateEvent(
		updateEventReq.Name,
		updateEventReq.Description,
		updateEventReq.IsDone,
		updateEventReq.SortId,
	)
}

// EventEntities2Reses 将检查事项实体列表转换为获取检查事项详情响应列表
// @param eList 检查事项实体列表
// @return 获取检查事项详情响应列表
func EventEntities2Reses(eventEntities []*entities.Event) types.ListEventRes {
	listRes := make([]*types.GetEventRes, 0, len(eventEntities))
	for _, e := range eventEntities {
		listRes = append(listRes, EventEntityToGetRes(e))
	}
	return listRes
}

// BatchUpdateEventReqToValueObjects 将批量更新检查事项请求体转换为批量更新检查事项值对象集合
// @param batchUpdateEventReq 批量更新检查事项请求体
// @return 批量更新检查事项值对象集合
// @return error 错误信息
func BatchUpdateEventReqToValueObjects(
	batchUpdateEventReq *types.BatchUpdateEventReq,
) ([]*valueobjects.BatchUpdateEvent, error) {
	batchVOs := make([]*valueobjects.BatchUpdateEvent, 0, len(batchUpdateEventReq.Events))
	for _, event := range batchUpdateEventReq.Events {
		id, err := strconv.ParseInt(event.Id, 10, 64)
		if err != nil {
			return nil, err
		}
		batchVO, err := valueobjects.NewBatchUpdateEvent(
			id,
			event.Name,
			event.Description,
			event.IsDone,
			event.SortId,
		)
		if err != nil {
			return nil, err
		}
		batchVOs = append(batchVOs, batchVO)
	}
	return batchVOs, nil
}
