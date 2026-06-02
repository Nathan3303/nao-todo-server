package task

import (
	"encoding/json"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"

	"gorm.io/gorm"
)

// 创建任务值对象转换为任务模型
// @param userId 用户ID
// @param createTaskValueObject 创建任务值对象
// @return 任务模型
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
		StartAt:        utils.TimePtr2SqlNullTime(createTaskValueObject.StartAt),
		EndAt:          utils.TimePtr2SqlNullTime(createTaskValueObject.EndAt),
		Tags:           createTaskValueObject.Tags,
		RemindAt:       utils.TimePtr2SqlNullTime(createTaskValueObject.RemindAt),
		RemindRepeat:   createTaskValueObject.RemindRepeat,
		RemindTime:     createTaskValueObject.RemindTime,
		RemindWeekdays: createTaskValueObject.RemindWeekdays,
	}
}

// 更新任务值对象转换为任务模型
// @param updateTaskValueObject 更新任务值对象
// @return 任务模型
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
	if updateTaskValueObject.StartAt != nil &&
		updateTaskValueObject.StartAt.ShouldUpdate() &&
		!updateTaskValueObject.StartAt.IsSetToNull() {
		m.StartAt = utils.ToSqlNullTime(updateTaskValueObject.StartAt)
	}
	if updateTaskValueObject.EndAt != nil &&
		updateTaskValueObject.EndAt.ShouldUpdate() &&
		!updateTaskValueObject.EndAt.IsSetToNull() {
		m.EndAt = utils.ToSqlNullTime(updateTaskValueObject.EndAt)
	}
	if updateTaskValueObject.ProjectId != nil {
		m.ProjectId = *updateTaskValueObject.ProjectId
	}
	if updateTaskValueObject.Tags != nil {
		m.Tags = updateTaskValueObject.Tags
	}
	if updateTaskValueObject.ArchivedAt != nil &&
		updateTaskValueObject.ArchivedAt.ShouldUpdate() &&
		!updateTaskValueObject.ArchivedAt.IsSetToNull() {
		m.ArchivedAt = utils.ToSqlNullTime(updateTaskValueObject.ArchivedAt)
	}
	if updateTaskValueObject.StarMarkAt != nil &&
		updateTaskValueObject.StarMarkAt.ShouldUpdate() &&
		!updateTaskValueObject.StarMarkAt.IsSetToNull() {
		m.StarMarkAt = utils.ToSqlNullTime(updateTaskValueObject.StarMarkAt)
	}
	if updateTaskValueObject.GivenUpAt != nil &&
		updateTaskValueObject.GivenUpAt.ShouldUpdate() &&
		!updateTaskValueObject.GivenUpAt.IsSetToNull() {
		m.GivenUpAt = utils.ToSqlNullTime(updateTaskValueObject.GivenUpAt)
	}
	if updateTaskValueObject.RemindAt != nil &&
		updateTaskValueObject.RemindAt.ShouldUpdate() &&
		!updateTaskValueObject.RemindAt.IsSetToNull() {
		m.RemindAt = utils.ToSqlNullTime(updateTaskValueObject.RemindAt)
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

// 更新任务值对象转换为任务映射
// @param updateTaskValueObject 更新任务值对象
// @return 任务映射
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
			updateMap["StartAt"] = utils.ToSqlNullTime(updateTaskValueObject.StartAt)
		}
	}
	if updateTaskValueObject.EndAt != nil && updateTaskValueObject.EndAt.ShouldUpdate() {
		if updateTaskValueObject.EndAt.IsSetToNull() {
			updateMap["EndAt"] = nil
		} else {
			updateMap["EndAt"] = utils.ToSqlNullTime(updateTaskValueObject.EndAt)
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
			updateMap["ArchivedAt"] = utils.ToSqlNullTime(updateTaskValueObject.ArchivedAt)
		}
	}
	if updateTaskValueObject.StarMarkAt != nil && updateTaskValueObject.StarMarkAt.ShouldUpdate() {
		if updateTaskValueObject.StarMarkAt.IsSetToNull() {
			updateMap["StarMarkAt"] = nil
		} else {
			updateMap["StarMarkAt"] = utils.ToSqlNullTime(updateTaskValueObject.StarMarkAt)
		}
	}
	if updateTaskValueObject.GivenUpAt != nil && updateTaskValueObject.GivenUpAt.ShouldUpdate() {
		if updateTaskValueObject.GivenUpAt.IsSetToNull() {
			updateMap["GivenUpAt"] = nil
		} else {
			updateMap["GivenUpAt"] = utils.ToSqlNullTime(updateTaskValueObject.GivenUpAt)
		}
	}
	if updateTaskValueObject.RemindAt != nil && updateTaskValueObject.RemindAt.ShouldUpdate() {
		if updateTaskValueObject.RemindAt.IsSetToNull() {
			updateMap["RemindAt"] = nil
		} else {
			updateMap["RemindAt"] = utils.ToSqlNullTime(updateTaskValueObject.RemindAt)
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

// 模型转换为任务实体
// @param m 任务模型
// @return 任务实体
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

// 分页值对象转换为分页范围
// @param pagination 分页值对象
// @return 分页范围
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

// 任务模型列表转换为任务实体列表
// @param mList 任务模型列表
// @return 任务实体列表
func TaskModels2Entities(mList []*models.Task) []*entities.Task {
	eList := make([]*entities.Task, 0, len(mList))
	for _, m := range mList {
		eList = append(eList, TaskModel2Entity(m))
	}
	return eList
}

// === CheckItem converters ===

func EventValueObjectToModel(vo *valueobjects.CreateCheckItem) *models.Event {
	return &models.Event{
		UserId:      vo.UserId,
		TaskId:      vo.TaskId,
		Name:        vo.Name,
		Description: vo.Description,
		SortId:      vo.SortId,
	}
}

func UpdateEventValueObjectToMap(vo *valueobjects.UpdateCheckItem) map[string]interface{} {
	m := make(map[string]interface{})
	if vo.Name != nil {
		m["Name"] = *vo.Name
	}
	if vo.Description != nil {
		m["Description"] = *vo.Description
	}
	if vo.IsDone != nil {
		m["IsDone"] = *vo.IsDone
	}
	if vo.SortId != nil {
		m["SortId"] = *vo.SortId
	}
	return m
}

func BatchUpdateEventValueObjectToMap(vo *valueobjects.BatchUpdateCheckItem) map[string]interface{} {
	m := make(map[string]interface{})
	if vo.Name != nil {
		m["Name"] = *vo.Name
	}
	if vo.Description != nil {
		m["Description"] = *vo.Description
	}
	if vo.IsDone != nil {
		m["IsDone"] = *vo.IsDone
	}
	if vo.SortId != nil {
		m["SortId"] = *vo.SortId
	}
	return m
}

func EventModel2Entity(m *models.Event) *entities.CheckItem {
	return &entities.CheckItem{
		Id:          m.ID,
		UserId:      m.UserId,
		TaskId:      m.TaskId,
		Name:        m.Name,
		Description: m.Description,
		IsDone:      m.IsDone,
		SortId:      m.SortId,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func EventModels2Entities(list []*models.Event) []*entities.CheckItem {
	result := make([]*entities.CheckItem, 0, len(list))
	for _, m := range list {
		result = append(result, EventModel2Entity(m))
	}
	return result
}

// === Comment converters ===

func CreateCommentValueObjectToModel(vo *valueobjects.CreateComment) *models.Comment {
	return &models.Comment{
		UserId:      vo.UserId,
		TaskId:      vo.TaskId,
		Content:     vo.Content,
		Attachments: vo.Attachments,
		IsTopUp:     vo.IsTopUp,
	}
}

func UpdateCommentValueObjectToMap(vo *valueobjects.UpdateComment) map[string]interface{} {
	m := make(map[string]interface{})
	if vo.Content != nil {
		m["Content"] = *vo.Content
	}
	if vo.Attachments != nil {
		if data, err := json.Marshal(vo.Attachments); err == nil {
			m["Attachments"] = string(data)
		}
	}
	if vo.IsTopUp != nil {
		m["IsTopUp"] = *vo.IsTopUp
	}
	return m
}

func CommentModel2Entity(m *models.Comment) *entities.Comment {
	return &entities.Comment{
		Id:          m.ID,
		UserId:      m.UserId,
		TaskId:      m.TaskId,
		Content:     m.Content,
		Attachments: m.Attachments,
		IsTopUp:     m.IsTopUp,
		Nickname:    m.Nickname,
		Avatar:      m.Avatar,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func CommentModels2Entities(list []*models.Comment) []*entities.Comment {
	result := make([]*entities.Comment, 0, len(list))
	for _, m := range list {
		result = append(result, CommentModel2Entity(m))
	}
	return result
}
