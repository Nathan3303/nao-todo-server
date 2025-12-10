package tag

import (
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/vo"
	"naotodoserver/infrastructure/persistence/models"
)

func TagEntity2Model(e *entities.Tag) *models.Tag {
	m := &models.Tag{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.Name = e.Name
	m.Description = e.Description
	m.Color = e.Color
	m.Preference = TagPreferenceVO2Model(e.Preference)
	return m
}

func TagPreferenceVO2Model(vo *vo.TagPreference) *models.TagPreference {
	if vo == nil {
		return nil
	}
	m := &models.TagPreference{}
	m.TagId = vo.TagId
	m.ViewType = vo.ViewType
	m.GetOptions = vo.GetOptions
	m.Columns = vo.Columns
	return m
}

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
	e.Preference = TagPreferenceModel2VO(m.Preference)
	return e
}

func TagPreferenceModel2VO(m *models.TagPreference) *vo.TagPreference {
	if m == nil {
		return nil
	}
	vo := &vo.TagPreference{}
	vo.Id = m.ID
	vo.TagId = m.TagId
	vo.ViewType = m.ViewType
	vo.GetOptions = m.GetOptions
	vo.Columns = m.Columns
	return vo
}

func TagModelList2EntityList(mList []*models.Tag) []*entities.Tag {
	eList := []*entities.Tag{}
	for _, m := range mList {
		eList = append(eList, TagModel2Entity(m))
	}
	return eList
}
