package comment

import (
	"encoding/json"
	"naotodoserver/domain/comment/entities"
	"naotodoserver/domain/comment/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// CreateCommentValueObjectToModel 创建评论值对象转换为评论模型
func CreateCommentValueObjectToModel(vo *valueobjects.CreateComment) *models.Comment {
	return &models.Comment{
		UserId:      vo.UserId,
		TaskId:      vo.TaskId,
		Content:     vo.Content,
		Attachments: vo.Attachments,
		IsTopUp:     vo.IsTopUp,
	}
}

// UpdateCommentValueObjectToModel 更新评论值对象转换为评论模型
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

// CommentModel2Entity 评论模型转换为评论实体
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

// CommentModels2Entities 评论模型列表转换为评论实体列表
func CommentModels2Entities(models []*models.Comment) []*entities.Comment {
	entities := make([]*entities.Comment, 0, len(models))
	for _, m := range models {
		entities = append(entities, CommentModel2Entity(m))
	}
	return entities
}
