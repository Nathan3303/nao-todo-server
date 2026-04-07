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
	m.TaskId = createEventValueObject.TaskId
	m.Name = createEventValueObject.Name
	m.Description = createEventValueObject.Description
	return m
}

// UpdateEventValueObjectToModel 更新事件值对象转换为事件模型
// @param updateEventValueObject 更新事件值对象
// @return 事件模型
func UpdateEventValueObjectToModel(updateEventValueObject *valueobjects.UpdateEvent) *models.Event {
	m := &models.Event{}
	m.Name = updateEventValueObject.Name
	m.Description = updateEventValueObject.Description
	m.IsDone = updateEventValueObject.IsDone
	m.SortId = updateEventValueObject.SortId
	return m
}

func EventModel2Entity(m *models.Event) *entities.Event {
	e := &entities.Event{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Name = m.Name
	e.Description = m.Description
	e.IsDone = m.IsDone
	e.SortId = m.SortId
	return e
}

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

func EventModels2Entities(mList []*models.Event) []*entities.Event {
	eList := make([]*entities.Event, 0, len(mList))
	for _, m := range mList {
		eList = append(eList, EventModel2Entity(m))
	}
	return eList
}
