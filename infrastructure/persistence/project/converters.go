package project

import (
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// CreateProjectValueObject2Model 创建项目 valueobject 转 model
// @param createProjectValueObject 项目创建值对象
// @return 项目模型
func CreateProjectValueObject2Model(
	createProjectValueObject *valueobjects.CreateProject,
) *models.Project {
	m := &models.Project{}
	m.UserId = createProjectValueObject.UserId
	m.Name = createProjectValueObject.Name
	m.Description = createProjectValueObject.Description
	return m
}

// UpdateProjectValueObject2Model 更新项目 valueobject 转 model
// @param updateProjectValueObject 项目更新值对象
// @return 项目模型
func UpdateProjectValueObject2Model(
	updateProjectValueObject *valueobjects.UpdateProject,
) *models.Project {
	m := &models.Project{}
	m.Name = *updateProjectValueObject.Name
	m.Description = *updateProjectValueObject.Description
	return m
}

// UpdateProjectValueObjectToMap 更新项目 valueobject 转 map
// - 用于更新项目时，将更新值对象转换为 map 格式以实现零值更新
// @param updateProjectValueObject 项目更新值对象
// @return map[string]any 项目更新值对象 map
func UpdateProjectValueObjectToMap(
	updateProjectValueObject *valueobjects.UpdateProject,
) map[string]interface{} {
	updateMap := make(map[string]interface{})
	if updateProjectValueObject.Name != nil {
		updateMap["Name"] = *updateProjectValueObject.Name
	}
	if updateProjectValueObject.Description != nil {
		updateMap["Description"] = *updateProjectValueObject.Description
	}
	// 输出更新 map 内容
	// println("更新 map 内容:")
	// println(updateMap)
	return updateMap
}

// Entity2Model 项目实体转模型
// @param e 项目实体
// @return 项目模型
func Entity2Model(e *entities.Project) *models.Project {
	m := &models.Project{}
	m.ID = e.Id
	m.UserId = e.UserId
	m.Name = e.Name
	m.Description = e.Description
	m.ArchivedAt = e.ArchivedAt
	m.DeactivedAt = e.DeactivedAt
	return m
}

// PreferenceVO2Model 项目偏好值对象转模型
// @param e 项目偏好值对象
// @return 项目偏好模型
func PreferenceVO2Model(e *entities.ProjectPreference) *models.ProjectPreference {
	m := &models.ProjectPreference{}
	m.UserId = e.UserId
	m.ProjectId = e.ProjectId
	m.ViewType = e.ViewType
	m.GetOptions = e.GetOptions
	m.Columns = e.Columns
	return m
}

// Model2Entity 项目模型转实体
// @param m 项目模型
// @return 项目实体
func Model2Entity(m *models.Project) *entities.Project {
	e := &entities.Project{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.Name = m.Name
	e.Description = m.Description
	e.ArchivedAt = m.ArchivedAt
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeactivedAt = m.DeactivedAt
	return e
}

// PreferenceModel2Entity 项目偏好模型转实体
// @param m 项目偏好模型
// @return 项目偏好实体
func PreferenceModel2Entity(m *models.ProjectPreference) *entities.ProjectPreference {
	if m == nil {
		return nil
	}
	e := &entities.ProjectPreference{}
	e.Id = m.ID
	e.UserId = m.UserId
	e.ProjectId = m.ProjectId
	e.ViewType = m.ViewType
	e.GetOptions = m.GetOptions
	e.Columns = m.Columns
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	return e
}

// Models2Entities 项目模型列表转实体列表
// @param ms 项目模型列表
// @return 项目实体列表
func Models2Entities(ms []*models.Project) []*entities.Project {
	es := make([]*entities.Project, 0, len(ms))
	for _, m := range ms {
		es = append(es, Model2Entity(m))
	}
	return es
}

// UpdateProjectPrefrenceValueObjectToModel 更新项目偏好值对象转模型
// @param userId 用户 ID
// @param projectId 项目 ID
// @param updateProjectPreferenceValueObject 更新项目偏好值对象
// @return 更新项目偏好模型
func UpdateProjectPrefrenceValueObjectToModel(
	userId int64,
	projectId int64,
	updateProjectPreferenceValueObject *valueobjects.SaveProjectPreference,
) *models.ProjectPreference {
	m := &models.ProjectPreference{}
	m.UserId = userId
	m.ProjectId = projectId
	m.ViewType = updateProjectPreferenceValueObject.ViewType
	m.GetOptions = updateProjectPreferenceValueObject.GetOptions
	m.Columns = updateProjectPreferenceValueObject.Columns
	return m
}
