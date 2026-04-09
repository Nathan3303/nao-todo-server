package tag

import (
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// CreateTagValueObjectToModel 创建标签值对象转换为标签模型
// @param userId 用户ID
// @param createTagValueObject 创建标签值对象
// @return models.Tag 标签模型
func CreateTagValueObjectToModel(
	userId int64,
	createTagValueObject *valueobjects.CreateTag,
) *models.Tag {
	return &models.Tag{
		UserId:      userId,
		Name:        createTagValueObject.Name,
		Description: createTagValueObject.Description,
		Color:       createTagValueObject.Color,
	}
}

// UpdateTagValueObjectToModel 更新标签值对象转换为标签模型
// @param userId 用户ID
// @param updateTagValueObject 更新标签值对象
// @return models.Tag 标签模型
func UpdateTagValueObjectToModel(
	userId int64,
	updateTagValueObject *valueobjects.UpdateTag,
) *models.Tag {
	return &models.Tag{
		UserId:      userId,
		Name:        updateTagValueObject.Name,
		Description: updateTagValueObject.Description,
		Color:       updateTagValueObject.Color,
	}
}

// TagEntity2Model 标签实体转换为标签模型
// @param e 标签实体
// @return models.Tag 标签模型
func TagEntity2Model(e *entities.Tag) *models.Tag {
	m := &models.Tag{}
	m.ID = e.Id
	m.Name = e.Name
	m.Description = e.Description
	m.Color = e.Color
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	return m
}

// TagPreferenceEntity2Model 标签偏好实体转换为标签偏好模型
// @param e 标签偏好实体
// @return models.TagPreference 标签偏好模型
func TagPreferenceEntity2Model(e *entities.TagPreference) *models.TagPreference {
	if e == nil {
		return nil
	}
	m := &models.TagPreference{}
	m.ID = e.Id
	m.TagId = e.TagId
	m.ViewType = e.ViewType
	m.GetOptions = e.GetOptions
	m.Columns = e.Columns
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	return m
}

// TagModel2Entity 标签模型转换为标签实体
// @param m 标签模型
// @return entities.Tag 标签实体
func TagModel2Entity(m *models.Tag) *entities.Tag {
	if m == nil {
		return nil
	}
	e := &entities.Tag{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.Name = m.Name
	e.Description = m.Description
	e.Color = m.Color
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return e
}

// TagPreferenceModel2VO 标签偏好模型转换为标签偏好值对象
// @param m 标签偏好模型
// @return entities.TagPreference 标签偏好实体
func TagPreferenceModel2Entity(m *models.TagPreference) *entities.TagPreference {
	if m == nil {
		return nil
	}
	e := &entities.TagPreference{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.TagId = m.TagId
	e.ViewType = m.ViewType
	e.GetOptions = m.GetOptions
	e.Columns = m.Columns
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return e
}

// TagModelList2EntityList 标签模型列表转换为标签实体列表
// @param mList 标签模型列表
// @return []*entities.Tag 标签实体列表
func TagModelList2EntityList(mList []*models.Tag) []*entities.Tag {
	eList := []*entities.Tag{}
	for _, m := range mList {
		eList = append(eList, TagModel2Entity(m))
	}
	return eList
}

// UpdateTagPreferenceValueObjectToModel 更新标签偏好值对象转换为标签偏好模型
// @param updateTagPreferenceValueObject 更新标签偏好值对象
// @return models.TagPreference 标签偏好模型
func UpdateTagPreferenceValueObjectToModel(
	userId int64,
	tagId int64,
	updateTagPreferenceValueObject *valueobjects.SaveTagPreference,
) *models.TagPreference {
	return &models.TagPreference{
		UserId:     userId,
		TagId:      tagId,
		ViewType:   updateTagPreferenceValueObject.ViewType,
		GetOptions: updateTagPreferenceValueObject.GetOptions,
		Columns:    updateTagPreferenceValueObject.Columns,
	}
}
