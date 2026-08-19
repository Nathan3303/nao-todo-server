package project

import (
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// CreateProjectValueObject2Model 创建项目 valueobject 转 model
// @param createProjectValueObject 项目创建值对象
// @return 项目模型
func CreateProjectValueObject2Model(
	createProjectValueObject *valueobjects.CreateProject,
) *models.Project {
	m := &models.Project{
		ModelBase: models.ModelBase{
			ID:        createProjectValueObject.Id,
			CreatedAt: createProjectValueObject.CreatedAt,
			UpdatedAt: createProjectValueObject.UpdatedAt,
			DeletedAt: gorm.DeletedAt(createProjectValueObject.DeletedAt.ToSqlNullTime()),
		},
	}
	m.UserId = createProjectValueObject.UserId
	m.Name = createProjectValueObject.Name
	m.Description = createProjectValueObject.Description
	m.SortId = createProjectValueObject.SortId
	return m
}

// CreateProjectVOToUpdateMap 创建项目值对象转换为全量更新映射（Upsert 覆盖用）
// 仅包含 Create VO 表达的字段，不触碰 archived_at/deactived_at 等列
// 客户端携带删除时间（本地墓碑）时写入 deleted_at；未携带时由 Upsert 兜底清空复活
func CreateProjectVOToUpdateMap(vo *valueobjects.CreateProject) map[string]any {
	updateMap := map[string]any{
		"Name":        vo.Name,
		"Description": vo.Description,
		"SortId":      vo.SortId,
	}
	if vo.DeletedAt.ShouldUpdate() && !vo.DeletedAt.IsSetToNull() {
		updateMap["deleted_at"] = vo.DeletedAt.ToSqlNullTime()
	}
	return updateMap
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
) map[string]any {
	updateMap := make(map[string]any)
	if updateProjectValueObject.Name != nil {
		updateMap["Name"] = *updateProjectValueObject.Name
	}
	if updateProjectValueObject.Description != nil {
		updateMap["Description"] = *updateProjectValueObject.Description
	}
	if updateProjectValueObject.SortId != nil {
		updateMap["SortId"] = *updateProjectValueObject.SortId
	}
	return updateMap
}

// Entity2Model 项目实体转模型
// @param e 项目实体
// @return 项目模型
func Entity2Model(e *entities.Project) *models.Project {
	m := &models.Project{}
	m.ID = e.Id
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = gorm.DeletedAt(e.DeletedAt.ToSqlNullTime())
	m.UserId = int64(e.UserId)
	m.Name = e.Name
	m.Description = e.Description
	m.ArchivedAt = e.ArchivedAt.ToSqlNullTime()
	m.DeactivedAt = e.DeactivedAt.ToSqlNullTime()
	m.SortId = e.SortId
	return m
}

// PreferenceVO2Model 项目偏好值对象转模型
// @param e 项目偏好值对象
// @return 项目偏好模型
func PreferenceVO2Model(e *entities.ProjectPreference) *models.ProjectPreference {
	m := &models.ProjectPreference{}
	m.ID = e.Id
	m.CreatedAt = e.CreatedAt
	m.UpdatedAt = e.UpdatedAt
	m.DeletedAt = gorm.DeletedAt(e.DeletedAt.ToSqlNullTime())
	m.UserId = int64(e.UserId)
	m.ProjectId = int64(e.ProjectId)
	m.ViewType = e.ViewType
	m.GetOptions = e.GetOptions
	m.Columns = e.Columns
	return m
}

// Model2Entity 项目模型转实体
// @param m 项目模型
// @return 项目实体
// Model2Entity函数需要更新时间字段和软删除字段的处理逻辑，与Entity2Model保持一致
func Model2Entity(m *models.Project) *entities.Project {
	e := &entities.Project{}
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.Name = m.Name
	e.Description = m.Description
	e.ArchivedAt = types.NewNullableTimeByTime(m.ArchivedAt.Time)
	e.DeactivedAt = types.NewNullableTimeByTime(m.DeactivedAt.Time)
	e.SortId = m.SortId
	return e
}

// PreferenceModel2Entity 项目偏好模型转实体
// @param m 项目偏好模型
// @return 项目偏好实体
func PreferenceModel2Entity(m *models.ProjectPreference) *entities.ProjectPreference {
	e := &entities.ProjectPreference{}
	e.Id = m.ID
	e.CreatedAt = m.CreatedAt
	e.UpdatedAt = m.UpdatedAt
	e.DeletedAt = types.NewNullableTimeByTime(m.DeletedAt.Time)
	e.UserId = types.UserID(m.UserId)
	e.ProjectId = types.ProjectID(m.ProjectId)
	e.ViewType = m.ViewType
	e.GetOptions = m.GetOptions
	e.Columns = m.Columns
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

// BatchUpdateProjectValueObjectToMap 批量更新项目值对象转 map
func BatchUpdateProjectValueObjectToMap(
	vo *valueobjects.BatchUpdateProject,
) map[string]any {
	updateMap := make(map[string]any)
	if vo.Name != nil {
		updateMap["Name"] = *vo.Name
	}
	if vo.Description != nil {
		updateMap["Description"] = *vo.Description
	}
	if vo.SortId != nil {
		updateMap["SortId"] = *vo.SortId
	}
	return updateMap
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
	m.ViewType = string(updateProjectPreferenceValueObject.ViewType)
	m.GetOptions = updateProjectPreferenceValueObject.GetOptions
	m.Columns = updateProjectPreferenceValueObject.Columns
	return m
}
