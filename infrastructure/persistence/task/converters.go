package task

import (
	"encoding/json"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// CreateTaskValueObjectToModel 创建任务值对象转换为任务模型
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
		StartAt:        createTaskValueObject.StartAt.ToSqlNullTime(),
		EndAt:          createTaskValueObject.EndAt.ToSqlNullTime(),
		Tags:           createTaskValueObject.Tags,
		RemindAt:       createTaskValueObject.RemindAt.ToSqlNullTime(),
		RemindRepeat:   createTaskValueObject.RemindRepeat,
		RemindTime:     createTaskValueObject.RemindTime,
		RemindWeekdays: createTaskValueObject.RemindWeekdays,
	}
}

// UpdateTaskValueObjectToModel 更新任务值对象转换为任务模型
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
	if updateTaskValueObject.StartAt.ShouldUpdate() &&
		!updateTaskValueObject.StartAt.IsSetToNull() {
		m.StartAt = updateTaskValueObject.StartAt.ToSqlNullTime()
	}
	if updateTaskValueObject.EndAt.ShouldUpdate() &&
		!updateTaskValueObject.EndAt.IsSetToNull() {
		m.EndAt = updateTaskValueObject.EndAt.ToSqlNullTime()
	}
	if updateTaskValueObject.ProjectId != nil {
		m.ProjectId = *updateTaskValueObject.ProjectId
	}
	if updateTaskValueObject.Tags != nil {
		m.Tags = updateTaskValueObject.Tags
	}
	if updateTaskValueObject.ArchivedAt.ShouldUpdate() &&
		!updateTaskValueObject.ArchivedAt.IsSetToNull() {
		m.ArchivedAt = updateTaskValueObject.ArchivedAt.ToSqlNullTime()
	}
	if updateTaskValueObject.StarMarkAt.ShouldUpdate() &&
		!updateTaskValueObject.StarMarkAt.IsSetToNull() {
		m.StarMarkAt = updateTaskValueObject.StarMarkAt.ToSqlNullTime()
	}
	if updateTaskValueObject.GivenUpAt.ShouldUpdate() &&
		!updateTaskValueObject.GivenUpAt.IsSetToNull() {
		m.GivenUpAt = updateTaskValueObject.GivenUpAt.ToSqlNullTime()
	}
	if updateTaskValueObject.RemindAt.ShouldUpdate() &&
		!updateTaskValueObject.RemindAt.IsSetToNull() {
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

// UpdateTaskValueObjectToMap 更新任务值对象转换为任务映射
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
	if updateTaskValueObject.StartAt.ShouldUpdate() {
		if updateTaskValueObject.StartAt.IsSetToNull() {
			updateMap["StartAt"] = nil
		} else {
			updateMap["StartAt"] = updateTaskValueObject.StartAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.EndAt.ShouldUpdate() {
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
	if updateTaskValueObject.ArchivedAt.ShouldUpdate() {
		if updateTaskValueObject.ArchivedAt.IsSetToNull() {
			updateMap["ArchivedAt"] = nil
		} else {
			updateMap["ArchivedAt"] = updateTaskValueObject.ArchivedAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.StarMarkAt.ShouldUpdate() {
		if updateTaskValueObject.StarMarkAt.IsSetToNull() {
			updateMap["StarMarkAt"] = nil
		} else {
			updateMap["StarMarkAt"] = updateTaskValueObject.StarMarkAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.GivenUpAt.ShouldUpdate() {
		if updateTaskValueObject.GivenUpAt.IsSetToNull() {
			updateMap["GivenUpAt"] = nil
		} else {
			updateMap["GivenUpAt"] = updateTaskValueObject.GivenUpAt.ToSqlNullTime()
		}
	}
	if updateTaskValueObject.RemindAt.ShouldUpdate() {
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
// @param m 任务模型
// @return 任务实体
func TaskModel2Entity(m *models.Task) *entities.Task {
	e := &entities.Task{}
	e.Id = m.ID
	e.UpdatedAt = m.UpdatedAt
	e.CreatedAt = m.CreatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = m.UserId
	e.ParentTaskId = m.ParentTaskId
	e.ProjectId = m.ProjectId
	e.Name = m.Name
	e.Description = m.Description
	e.State = m.State
	e.Priority = m.Priority
	e.StartAt = types.NewNullableTimeByTime(m.StartAt.Time)
	e.EndAt = types.NewNullableTimeByTime(m.EndAt.Time)
	e.ArchivedAt = types.NewNullableTimeByTime(m.ArchivedAt.Time)
	e.StarMarkAt = types.NewNullableTimeByTime(m.StarMarkAt.Time)
	e.GivenUpAt = types.NewNullableTimeByTime(m.GivenUpAt.Time)
	e.RemindAt = types.NewNullableTimeByTime(m.RemindAt.Time)
	e.RemindRepeat = m.RemindRepeat
	e.RemindTime = m.RemindTime
	e.RemindWeekdays = m.RemindWeekdays
	e.Tags = m.Tags
	return e
}

// PaginationVO2Scopes 分页值对象转换为分页范围
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

// TaskModels2Entities 任务模型列表转换为任务实体列表
// @param mList 任务模型列表
// @return 任务实体列表
func TaskModels2Entities(mList []*models.Task) []*entities.Task {
	eList := make([]*entities.Task, 0, len(mList))
	for _, m := range mList {
		eList = append(eList, TaskModel2Entity(m))
	}
	return eList
}

// --- 检查项相关 ---

// TaskCheckItemValueObjectToModel 创建检查项值对象转换为检查项模型
func TaskCheckItemValueObjectToModel(vo *valueobjects.CreateTaskCheckItem) *models.TaskCheckItem {
	return &models.TaskCheckItem{
		UserId:      vo.UserId,
		TaskId:      vo.TaskId,
		Name:        vo.Name,
		Description: vo.Description,
		SortId:      vo.SortId,
	}
}

// UpdateTaskCheckItemValueObjectToMap 更新检查项值对象转换为更新映射
func UpdateTaskCheckItemValueObjectToMap(
	vo *valueobjects.UpdateTaskCheckItem,
) map[string]interface{} {
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

// BatchUpdateTaskCheckItemValueObjectToMap 批量更新检查项值对象转换为批量更新映射
func BatchUpdateTaskCheckItemValueObjectToMap(
	vo *valueobjects.BatchUpdateTaskCheckItem,
) map[string]interface{} {
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

// TaskCheckItemModel2Entity 转换为任务检查项实体
// @param m 任务检查项模型
// @return 任务检查项实体
func TaskCheckItemModel2Entity(m *models.TaskCheckItem) *entities.TaskCheckItem {
	var e entities.TaskCheckItem
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Name = m.Name
	e.Description = m.Description
	e.IsDone = m.IsDone
	e.SortId = m.SortId
	return &e
}

// TaskCheckItemModels2Entities 转换为任务检查项实体列表
// @param list 任务检查项模型列表
// @return 任务检查项实体列表
func TaskCheckItemModels2Entities(list []*models.TaskCheckItem) []*entities.TaskCheckItem {
	result := make([]*entities.TaskCheckItem, 0, len(list))
	for _, m := range list {
		result = append(result, TaskCheckItemModel2Entity(m))
	}
	return result
}

// --- Task Comment相关 ---

// CreateTaskCommentValueObjectToModel 创建任务评论值对象转换为任务评论模型
func CreateTaskCommentValueObjectToModel(vo *valueobjects.CreateTaskComment) *models.TaskComment {
	return &models.TaskComment{
		UserId:      vo.UserId,
		TaskId:      vo.TaskId,
		Content:     vo.Content,
		Attachments: vo.Attachments,
		IsTopUp:     vo.IsTopUp,
	}
}

// UpdateTaskCommentValueObjectToMap 更新任务评论值对象转换为更新映射
func UpdateTaskCommentValueObjectToMap(vo *valueobjects.UpdateTaskComment) map[string]interface{} {
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

// TaskCommentModel2Entity 转换为任务评论实体
// @param m 任务评论模型
// @return 任务评论实体
func TaskCommentModel2Entity(m *models.TaskComment) *entities.TaskComment {
	var e entities.TaskComment
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Content = m.Content
	e.Attachments = m.Attachments
	e.IsTopUp = m.IsTopUp
	e.Nickname = m.Nickname
	e.Avatar = m.Avatar
	return &e
}

// TaskCommentModels2Entities 转换为任务评论实体列表
// @param list 任务评论模型列表
// @return 任务评论实体列表
func TaskCommentModels2Entities(list []*models.TaskComment) []*entities.TaskComment {
	result := make([]*entities.TaskComment, 0, len(list))
	for _, m := range list {
		result = append(result, TaskCommentModel2Entity(m))
	}
	return result
}
