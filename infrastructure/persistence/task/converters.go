package task

import (
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/vo"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"

	"gorm.io/gorm"
)

func TaskEntity2Model(e *entities.Task) *models.Task {
	m := &models.Task{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.ProjectId = e.ProjectId
	m.Name = e.Name
	m.Description = e.Description
	m.State = e.State
	m.Priority = e.Priority
	m.StartAt = utils.TimePtr2SqlNullTime(e.StartAt)
	m.EndAt = utils.TimePtr2SqlNullTime(e.EndAt)
	m.ArchivedAt = utils.TimePtr2SqlNullTime(e.ArchivedAt)
	m.StarMarkAt = utils.TimePtr2SqlNullTime(e.StarMarkAt)
	m.GivenUpAt = utils.TimePtr2SqlNullTime(e.GivenUpAt)
	m.Tags = e.Tags
	m.UpdatedAt = e.UpdatedAt
	m.CreatedAt = e.CreatedAt
	m.DeletedAt = gorm.DeletedAt{Time: e.DeletedAt}
	return m
}

func TaskModel2Entity(m *models.Task) *entities.Task {
	e := &entities.Task{}
	e.Id = m.ID
	e.UserId = m.UserId
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
	e.Tags = m.Tags
	e.UpdatedAt = m.UpdatedAt
	e.CreatedAt = m.CreatedAt
	e.DeletedAt = m.DeletedAt.Time
	return e
}

func PaginationVO2Scopes(pagination *vo.Pagination) func(db *gorm.DB) *gorm.DB {
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
