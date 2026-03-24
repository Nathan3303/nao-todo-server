package project

import (
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/vo"
	"naotodoserver/infrastructure/persistence/models"
)

func Entity2Model(e *entities.Project) *models.Project {
	m := &models.Project{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.Name = e.Name
	m.Description = e.Description
	m.ArchivedAt = e.ArchivedAt
	// m.Preference = PreferenceVO2Model(e.Preference)
	return m
}

func PreferenceVO2Model(e *vo.ProjectPreference) *models.ProjectPreference {
	m := &models.ProjectPreference{}
	m.UserId = e.UserId
	m.ProjectId = e.ProjectId
	m.ViewType = e.ViewType
	m.GetOptions = e.GetOptions
	m.Columns = e.Columns
	return m
}

func Model2Entity(m *models.Project) *entities.Project {
	e := &entities.Project{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.Name = m.Name
	e.Description = m.Description
	e.ArchivedAt = m.ArchivedAt
	e.Preference = PreferenceModel2VO(m.Preference)
	return e
}

func PreferenceModel2VO(m *models.ProjectPreference) *vo.ProjectPreference {
	if m == nil {
		return nil
	}
	vo := &vo.ProjectPreference{}
	vo.UserId = m.UserId
	vo.ProjectId = m.ProjectId
	vo.ViewType = m.ViewType
	vo.GetOptions = m.GetOptions
	vo.Columns = m.Columns
	return vo
}

func Models2Entities(ms []*models.Project) []*entities.Project {
	es := make([]*entities.Project, 0, len(ms))
	for _, m := range ms {
		es = append(es, Model2Entity(m))
	}
	return es
}
