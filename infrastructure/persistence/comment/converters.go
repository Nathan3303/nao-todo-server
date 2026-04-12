package comment

import (
	"encoding/json"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// CreateCommentValueObjectToModel 创建评论值对象转换为评论模型
// @param vo 创建评论值对象
// @return 评论模型
func CreateCommentValueObjectToModel(vo *valueobjects.CreateComment) *models.Comment {
	m := &models.Comment{}
	m.UserId = vo.UserId
	m.TaskId = vo.TaskId
	m.Content = vo.Content
	m.Attachments = vo.Attachments
	m.IsTopUp = vo.IsTopUp
	return m
}

// UpdateCommentValueObjectToModel 更新评论值对象转换为评论模型
// @param vo 更新评论值对象
// @return 评论模型
func UpdateCommentValueObjectToModel(vo *valueobjects.UpdateComment) *models.Comment {
	m := &models.Comment{}
	if vo.Content != nil {
		m.Content = *vo.Content
	}
	if vo.Attachments != nil {
		m.Attachments = *vo.Attachments
	}
	if vo.IsTopUp != nil {
		m.IsTopUp = *vo.IsTopUp
	}
	return m
}

// UpdateCommentValueObjectToMap 更新评论值对象转换为 map
// - 用于更新评论时，将更新值对象转换为 map 格式以实现零值更新
// @param vo 更新评论值对象
// @return map[string]any 评论更新值对象 map
func UpdateCommentValueObjectToMap(vo *valueobjects.UpdateComment) map[string]interface{} {
	updateMap := make(map[string]interface{})
	if vo.Content != nil {
		updateMap["Content"] = *vo.Content
	}
	if vo.Attachments != nil {
		attachmentsJson, err := json.Marshal(vo.Attachments)
		if err == nil {
			updateMap["Attachments"] = string(attachmentsJson)
		}
	}
	if vo.IsTopUp != nil {
		updateMap["IsTopUp"] = *vo.IsTopUp
	}
	return updateMap
}

// CommentEntity2Model 评论实体转换为评论模型
// @param e 评论实体
// @return 评论模型
func CommentEntity2Model(e *entities.Comment) *models.Comment {
	m := &models.Comment{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.TaskId = e.TaskId
	m.Content = e.Content
	m.Attachments = e.Attachments
	m.IsTopUp = e.IsTopUp
	m.CommentUser = CommentUserEntity2Model(e.CommentUser)
	return m
}

// CommentUserEntity2Model 评论用户实体转换为评论用户模型
// @param e 评论用户实体
// @return 评论用户模型
func CommentUserEntity2Model(e *entities.CommentUser) *models.CommentUser {
	if e == nil {
		return nil
	}
	m := &models.CommentUser{}
	m.ID = e.Id
	m.CommentId = e.CommentId
	m.Nickname = e.Nickname
	m.Avatar = e.Avatar
	return m
}

// CommentModel2Entity 评论模型转换为评论实体
// @param m 评论模型
// @return 评论实体
func CommentModel2Entity(m *models.Comment) *entities.Comment {
	e := &entities.Comment{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.TaskId = m.TaskId
	e.Content = m.Content
	e.Attachments = m.Attachments
	e.IsTopUp = m.IsTopUp
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.CommentUser = CommentUserModel2Entity(m.CommentUser)
	return e
}

// CommentUserModel2Entity 评论用户模型转换为评论用户实体
// @param m 评论用户模型
// @return 评论用户实体
func CommentUserModel2Entity(m *models.CommentUser) *entities.CommentUser {
	e := &entities.CommentUser{}
	e.Id = m.ID
	e.CommentId = m.CommentId
	e.Nickname = m.Nickname
	e.Avatar = m.Avatar
	return e
}

// CommentModels2Entities 评论模型列表转换为评论实体列表
// @param models 评论模型列表
// @return 评论实体列表
func CommentModels2Entities(models []*models.Comment) []*entities.Comment {
	entities := make([]*entities.Comment, 0, len(models))
	for _, m := range models {
		entities = append(entities, CommentModel2Entity(m))
	}
	return entities
}
