package comment

import (
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/vo"
	"naotodoserver/infrastructure/persistence/models"
)

func CommentEntity2Model(e *entities.Comment) *models.Comment {
	m := &models.Comment{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.TaskId = e.TaskId
	m.Content = e.Content
	m.Attachments = e.Attachments
	m.IsTopUp = e.IsTopUp
	m.CommentUser = CommentUserVO2Model(e.CommentUser)
	return m
}

func CommentUserVO2Model(vo *vo.CommentUser) *models.CommentUser {
	if vo == nil {
		return nil
	}
	m := &models.CommentUser{}
	m.ID = vo.Id
	m.CommentId = vo.CommentId
	m.Nickname = vo.Nickname
	m.Avatar = vo.Avatar
	return m
}

func CommentModel2Entity(m *models.Comment) *entities.Comment {
	e := &entities.Comment{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Content = m.Content
	e.Attachments = m.Attachments
	e.IsTopUp = m.IsTopUp
	e.CreatedAt = &m.CreatedAt
	e.CommentUser = CommentUserModel2VO(m.CommentUser)
	return e
}

func CommentUserModel2VO(m *models.CommentUser) *vo.CommentUser {
	vo := &vo.CommentUser{}
	vo.Id = m.ID
	vo.CommentId = m.CommentId
	vo.Nickname = m.Nickname
	vo.Avatar = m.Avatar
	return vo
}

func CommentModels2Entities(models []*models.Comment) []*entities.Comment {
	entities := make([]*entities.Comment, 0, len(models))
	for _, m := range models {
		entities = append(entities, CommentModel2Entity(m))
	}
	return entities
}
