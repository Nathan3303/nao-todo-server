package event

import (
	"naotodoserver/domain/event/entities"
	"naotodoserver/domain/event/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// CreateEventValueObjectToModel 创建事件值对象转换为事件模型
// @param createEventValueObject 创建事件值对象
// @return 事件模型
func CreateEventValueObjectToModel(createEventValueObject *valueobjects.CreateEvent) *models.Event {
	m := &models.Event{}
	m.UserId = createEventValueObject.UserId
	m.TaskId = createEventValueObject.TaskId
	m.Name = createEventValueObject.Name
	m.Description = createEventValueObject.Description
	m.SortId = createEventValueObject.SortId
	return m
}

// UpdateEventValueObjectToModel 更新事件值对象转换为事件模型
// @param updateEventValueObject 更新事件值对象
// @return 事件模型
func UpdateEventValueObjectToModel(updateEventValueObject *valueobjects.UpdateEvent) *models.Event {
	m := &models.Event{}
	if updateEventValueObject.Name != nil {
		m.Name = *updateEventValueObject.Name
	}
	if updateEventValueObject.Description != nil {
		m.Description = *updateEventValueObject.Description
	}
	if updateEventValueObject.IsDone != nil {
		m.IsDone = *updateEventValueObject.IsDone
	}
	if updateEventValueObject.SortId != nil {
		m.SortId = *updateEventValueObject.SortId
	}
	return m
}

// UpdateEventValueObjectToMap 更新事件值对象转换为 map
// - 用于更新事件时，将更新值对象转换为 map 格式以实现零值更新
// @param updateEventValueObject 更新事件值对象
// @return map[string]any 事件更新值对象 map
func UpdateEventValueObjectToMap(
	updateEventValueObject *valueobjects.UpdateEvent,
) map[string]interface{} {
	updateMap := make(map[string]interface{})
	if updateEventValueObject.Name != nil {
		updateMap["Name"] = *updateEventValueObject.Name
	}
	if updateEventValueObject.Description != nil {
		updateMap["Description"] = *updateEventValueObject.Description
	}
	if updateEventValueObject.IsDone != nil {
		updateMap["IsDone"] = *updateEventValueObject.IsDone
	}
	if updateEventValueObject.SortId != nil {
		updateMap["SortId"] = *updateEventValueObject.SortId
	}
	return updateMap
}

// EventModel2Entity 事件模型转换为事件实体
// @param m models.Event 事件模型
// @return entities.Event 事件实体
func EventModel2Entity(m *models.Event) *entities.Event {
	e := &entities.Event{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Name = m.Name
	e.Description = m.Description
	e.IsDone = m.IsDone
	e.SortId = m.SortId
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return e
}

// EventEntity2Model 事件实体转换为事件模型
// @param e entities.Event 事件实体
// @return models.Event 事件模型
func EventEntity2Model(e *entities.Event) *models.Event {
	m := &models.Event{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.TaskId = e.TaskId
	m.Name = e.Name
	m.Description = e.Description
	m.IsDone = e.IsDone
	m.SortId = e.SortId
	return m
}

// EventModels2Entities 事件模型列表转换为事件实体列表
// @param mList []*models.Event 事件模型列表
// @return []*entities.Event 事件实体列表
func EventModels2Entities(mList []*models.Event) []*entities.Event {
	eList := make([]*entities.Event, 0, len(mList))
	for _, m := range mList {
		eList = append(eList, EventModel2Entity(m))
	}
	return eList
}
