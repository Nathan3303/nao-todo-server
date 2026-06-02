package checkitem

import (
	"naotodoserver/domain/checkitem/entities"
	"naotodoserver/domain/checkitem/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
)

// EventEntityToGetRes 将检查事项实体转换为获取检查事项详情响应
// @param e 检查事项实体
// @return 获取检查事项详情响应
func EventEntityToGetRes(e *entities.CheckItem) *types.GetCheckItemRes {
	return &types.GetCheckItemRes{
		Id:          strconv.FormatInt(e.Id, 10),
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		Name:        e.Name,
		Description: e.Description,
		IsDone:      e.IsDone,
		SortId:      e.SortId,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// CreateEventReqToValueObject 将创建检查事项请求体转换为创建检查事项值对象
// @param userId 用户ID
// @param createReq 创建检查事项请求体
// @return 创建检查事项值对象
// @return error 错误信息
func CreateEventReqToValueObject(
	userId int64,
	createReq *types.CreateCheckItemReq,
) (*valueobjects.CreateCheckItem, error) {
	taskId, err := strconv.ParseInt(createReq.TaskId, 10, 64)
	if err != nil {
		return nil, err
	}
	return valueobjects.NewCreateCheckItem(
		userId,
		taskId,
		createReq.Name,
		createReq.Description,
	)
}

// EventEntityToCreateRes 将检查事项实体转换为创建检查事项响应
// @param e 检查事项实体
// @return 创建检查事项响应
func EventEntityToCreateRes(e *entities.CheckItem) *types.CreateCheckItemRes {
	return &types.CreateCheckItemRes{
		Id:          strconv.FormatInt(e.Id, 10),
		TaskId:      strconv.FormatInt(e.TaskId, 10),
		Name:        e.Name,
		Description: e.Description,
		IsDone:      e.IsDone,
		SortId:      e.SortId,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// UpdateEventReqToValueObject 将更新检查事项请求体转换为更新检查事项值对象
// @param updateReq 更新检查事项请求体
// @return 更新检查事项值对象
// @return error 错误信息
func UpdateEventReqToValueObject(
	updateReq *types.UpdateCheckItemReq,
) (*valueobjects.UpdateCheckItem, error) {
	return valueobjects.NewUpdateCheckItem(
		updateReq.Name,
		updateReq.Description,
		updateReq.IsDone,
		updateReq.SortId,
	)
}

// EventEntities2Reses 将检查事项实体列表转换为获取检查事项详情响应列表
// @param eList 检查事项实体列表
// @return 获取检查事项详情响应列表
func EventEntities2Reses(items []*entities.CheckItem) types.ListCheckItemRes {
	listRes := make([]*types.GetCheckItemRes, 0, len(items))
	for _, e := range items {
		listRes = append(listRes, EventEntityToGetRes(e))
	}
	return listRes
}

// BatchUpdateEventReqToValueObjects 将批量更新检查事项请求体转换为批量更新检查事项值对象集合
// @param batchReq 批量更新检查事项请求体
// @return 批量更新检查事项值对象集合
// @return error 错误信息
func BatchUpdateEventReqToValueObjects(
	batchReq *types.BatchUpdateCheckItemReq,
) ([]*valueobjects.BatchUpdateCheckItem, error) {
	batchVOs := make([]*valueobjects.BatchUpdateCheckItem, 0, len(batchReq.Events))
	for _, event := range batchReq.Events {
		id, err := strconv.ParseInt(event.Id, 10, 64)
		if err != nil {
			return nil, err
		}
		batchVO, err := valueobjects.NewBatchUpdateCheckItem(
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
