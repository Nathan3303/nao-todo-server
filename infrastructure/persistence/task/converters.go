package task

import (
	"encoding/json"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"

	"gorm.io/gorm"
)

func CreateTaskValueObjectToModel(
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) *models.Task {
	return &models.Task{
		UserId:         userId,
		ProjectId:      createTaskValueObject.ProjectId,
		Name:           createTaskValueObject.Name,
		Description:    createTaskValueObject.Description,
		State:          createTaskValueObject.State,
		Priority:       createTaskValueObject.Priority,
		StartAt:        createTaskValueObject.StartAt,
		EndAt:          createTaskValueObject.EndAt,
		Tags:           createTaskValueObject.Tags,
		RemindAt:       createTaskValueObject.RemindAt,
		RemindRepeat:   createTaskValueObject.RemindRepeat,
		RemindTime:     createTaskValueObject.RemindTime,
		RemindWeekdays: createTaskValueObject.RemindWeekdays,
	}
}

func UpdateTaskValueObjectToModel(
	updateTaskValueObject *valueobjects.UpdateTask,
) *models.Task {
	m := &models.Task{}
	if updateTaskValueObject.UserId == 0 {
		m.UserId = updateTaskValueObject.UserId
	}
	if updateTaskValueObject.ParentTaskId != nil {
		m.ParentTaskId = *updateTaskValueObject.ParentTaskId
	}
	if updateTaskValueObject.Name != nil {
		m.Name = *updateTaskValueObject.Name
	}
	if updateTaskValueObject.Description != nil {
		m.Description = *updateTaskValueObject.Description
	}
	if updateTaskValueObject.State != nil {
		m.State = *updateTaskValueObject.State
	}
	if updateTaskValueObject.Priority != nil {
		m.Priority = *updateTaskValueObject.Priority
	}
	if updateTaskValueObject.StartAt != nil && updateTaskValueObject.StartAt.ShouldUpdate() && !updateTaskValueObject.StartAt.IsSetToNull() {
		m.StartAt = updateTaskValueObject.StartAt.ToSqlNullTime()
	}
	if updateTaskValueObject.EndAt != nil && updateTaskValueObject.EndAt.ShouldUpdate() && !updateTaskValueObject.EndAt.IsSetToNull() {
		m.EndAt = updateTaskValueObject.EndAt.ToSqlNullTime()
	}
	if updateTaskValueObject.ProjectId != nil {
		m.ProjectId = *updateTaskValueObject.ProjectId
	}
	if updateTaskValueObject.Tags != nil {
		m.Tags = updateTaskValueObject.Tags
	}
	if updateTaskValueObject.ArchivedAt != nil && updateTaskValueObject.ArchivedAt.ShouldUpdate() && !updateTaskValueObject.ArchivedAt.IsSetToNull() {
		m.ArchivedAt = updateTaskValueObject.ArchivedAt.ToSqlNullTime()
	}
	if updateTaskValueObject.StarMarkAt != nil && updateTaskValueObject.StarMarkAt.ShouldUpdate() && !updateTaskValueObject.StarMarkAt.IsSetToNull() {
		m.StarMarkAt = updateTaskValueObject.StarMarkAt.ToSqlNullTime()
	}
	if updateTaskValueObject.GivenUpAt != nil && updateTaskValueObject.GivenUpAt.ShouldUpdate() && !updateTaskValueObject.GivenUpAt.IsSetToNull() {
		m.GivenUpAt = updateTaskValueObject.GivenUpAt.ToSqlNullTime()
	}
	if updateTaskValueObject.RemindAt != nil && updateTaskValueObject.RemindAt.ShouldUpdate() && !updateTaskValueObject.RemindAt.IsSetToNull() {
		m.RemindAt = updateTaskValueObject.RemindAt.ToSqlNullTime()
	}
	if updateTaskValueObject.RemindRepeat != nil {
		m.RemindRepeat = *updateTaskValueObject.RemindRepeat
	}
	if updateTaskValueObject.RemindTime != nil {
		m.RemindTime = *updateTaskValueObject.RemindTime
	}
	if updateTaskValueObject.RemindWeekdays != nil {
		m.RemindWeekdays = *updateTaskValueObject.RemindWeekdays
	}
	return m
}

func UpdateTaskValueObjectToMap(
	updateTaskValueObject *valueobjects.UpdateTask,
) map[string]interface{} {
	updateMap := make(map[string]interface{})
	if updateTaskValueObject.UserId != 0 {
		updateMap["UserId"] = updateTaskValueObject.UserId
	}
	if updateTaskValueObject.ParentTaskId != nil {
		updateMap["ParentTaskId"] = *updateTaskValueObject.ParentTaskId
	}
	if updateTaskValueObject.Name != nil {
		updateMap["Name"] = *updateTaskValueObject.Name
	}
	if updateTaskValueObject.Description != nil {
		updateMap["Description"] = *updateTaskValueObject.Description
	}
	if updateTaskValueObject.State != nil {
		updateMap["State"] = *updateTaskValueObject.State
	}
	if updateTaskValueObject.Priority != nil {
		updateMap["Priority"] = *updateTaskValueObject.Priority
	}
	if updateTaskValueObject.StartAt != nil && updateTaskValueObject.StartAt.ShouldUpdate() {
		if updateTaskValueObject.StartAt.IsSetToNull() {
			updateMap["StartAt"] = nil
		} else {
			updateMap["StartAt"] = updateTaskValueObject.StartAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.EndAt != nil && updateTaskValueObject.EndAt.ShouldUpdate() {
		if updateTaskValueObject.EndAt.IsSetToNull() {
			updateMap["EndAt"] = nil
		} else {
			updateMap["EndAt"] = updateTaskValueObject.EndAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.ProjectId != nil {
		updateMap["ProjectId"] = *updateTaskValueObject.ProjectId
	}
	if updateTaskValueObject.Tags != nil {
		tagsJSON, err := json.Marshal(updateTaskValueObject.Tags)
		if err == nil {
			updateMap["Tags"] = string(tagsJSON)
		}
	}
	if updateTaskValueObject.ArchivedAt != nil && updateTaskValueObject.ArchivedAt.ShouldUpdate() {
		if updateTaskValueObject.ArchivedAt.IsSetToNull() {
			updateMap["ArchivedAt"] = nil
		} else {
			updateMap["ArchivedAt"] = updateTaskValueObject.ArchivedAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.StarMarkAt != nil && updateTaskValueObject.StarMarkAt.ShouldUpdate() {
		if updateTaskValueObject.StarMarkAt.IsSetToNull() {
			updateMap["StarMarkAt"] = nil
		} else {
			updateMap["StarMarkAt"] = updateTaskValueObject.StarMarkAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.GivenUpAt != nil && updateTaskValueObject.GivenUpAt.ShouldUpdate() {
		if updateTaskValueObject.GivenUpAt.IsSetToNull() {
			updateMap["GivenUpAt"] = nil
		} else {
			updateMap["GivenUpAt"] = updateTaskValueObject.GivenUpAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.RemindAt != nil && updateTaskValueObject.RemindAt.ShouldUpdate() {
		if updateTaskValueObject.RemindAt.IsSetToNull() {
			updateMap["RemindAt"] = nil
		} else {
			updateMap["RemindAt"] = updateTaskValueObject.RemindAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.RemindRepeat != nil {
		updateMap["RemindRepeat"] = *updateTaskValueObject.RemindRepeat
	}
	if updateTaskValueObject.RemindTime != nil {
		updateMap["RemindTime"] = *updateTaskValueObject.RemindTime
	}
	if updateTaskValueObject.RemindWeekdays != nil {
		updateMap["RemindWeekdays"] = *updateTaskValueObject.RemindWeekdays
	}
	return updateMap
}

// TaskModel2Entity 模型转换为任务实体
func TaskModel2Entity(m *models.Task) *entities.Task {
	e := &entities.Task{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.ParentTaskId = m.ParentTaskId
	e.ProjectId = m.ProjectId
	e.Name = m.Name
	e.Description = m.Description
	e.State = m.State
	e.Priority = m.Priority
	e.StartAt = utils.SqlNullTime2TimePtr(m.StartAt)
	e.EndAt = utils.SqlNullTime2TimePtr(m.EndAt)
	e.ArchivedAt = utils.SqlNullTime2TimePtr(m.ArchivedAt)
	e.StarMarkAt = utils.SqlNullTime2TimePtr(m.StarMarkAt)
	e.GivenUpAt = utils.SqlNullTime2TimePtr(m.GivenUpAt)
	e.RemindAt = utils.SqlNullTime2TimePtr(m.RemindAt)
	e.RemindRepeat = m.RemindRepeat
	e.RemindTime = m.RemindTime
	e.RemindWeekdays = m.RemindWeekdays
	e.Tags = m.Tags
	e.UpdatedAt = m.UpdatedAt
	e.CreatedAt = m.CreatedAt
	e.DeletedAt = m.DeletedAt.Time
	return e
}

func PaginationVO2Scopes(pagination *valueobjects.Pagination) func(db *gorm.DB) *gorm.DB {
	if pagination.Page <= 0 {
		pagination.Page = 1
	}
	if pagination.Limit <= 0 {
		pagination.Limit = 10
	}
	offset := (pagination.Page - 1) * pagination.Limit
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(pagination.Limit)
	}
}

func TaskModels2Entities(mList []*models.Task) []*entities.Task {
	eList := make([]*entities.Task, 0, len(mList))
	for _, m := range mList {
		eList = append(eList, TaskModel2Entity(m))
	}
	return eList
}
