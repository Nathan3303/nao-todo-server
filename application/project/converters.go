package project

import (
	"naotodoserver/application/idutil"
	"naotodoserver/application/project/dto"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"time"
)

// CreateProjectReqToValueObject 创建任务清单请求体转换值对象
// @param userId 用户 ID
// @param req 创建任务清单请求体
// @return *valueobjects.CreateProject 创建任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func CreateProjectReqToValueObject(
	userId int64,
	req *dto.CreateProjectReq,
) (*valueobjects.CreateProject, error) {
	createProjectValueObject, err := valueobjects.NewCreateProject(
		userId,
		req.Name,
		req.Description,
	)
	if err != nil {
		return nil, err
	}
	return createProjectValueObject, nil
}

// ProjectEntityToGetRes 任务清单实体转换响应体
// @param projectEntity 任务清单实体
// @return *dto.GetProjectRes 获取任务清单响应体
func ProjectEntityToGetRes(projectEntity *entities.Project) *dto.GetProjectRes {
	var res dto.GetProjectRes
	res.Id = idutil.FormatID(projectEntity.Id)
	res.CreatedAt = projectEntity.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = projectEntity.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = projectEntity.DeletedAt.ToString(time.RFC3339)
	res.Name = projectEntity.Name
	res.Description = projectEntity.Description
	res.SortId = projectEntity.SortId
	res.ArchivedAt = projectEntity.ArchivedAt.ToString(time.RFC3339)
	res.DeactivedAt = projectEntity.DeactivedAt.ToString(time.RFC3339)
	return &res
}

// UpdateProjectReqToValueObject 更新任务清单请求体转换值对象
// @param updateProjectReq 更新任务清单请求体
// @return *valueobjects.UpdateProject 更新任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func UpdateProjectReqToValueObject(
	updateProjectReq *dto.UpdateProjectReq,
) (*valueobjects.UpdateProject, error) {
	updateProjectValueObject, err := valueobjects.NewUpdateProject(
		updateProjectReq.Name,
		updateProjectReq.Description,
		updateProjectReq.SortId,
	)
	if err != nil {
		return nil, err
	}
	return updateProjectValueObject, nil
}

// EntitiesToGetResList 任务清单实体列表转换响应体列表
// @param projectEntities 任务清单实体列表
// @return []*dto.GetProjectRes 任务清单响应体列表
func EntitiesToGetResList(projectEntities []*entities.Project) []*dto.GetProjectRes {
	getResList := make([]*dto.GetProjectRes, 0, len(projectEntities))
	for _, projectEntity := range projectEntities {
		getResList = append(getResList, ProjectEntityToGetRes(projectEntity))
	}
	return getResList
}

// BatchUpdateProjectReqToValueObjects 批量更新任务清单请求体转换值对象
// @param req 批量更新任务清单请求体
// @return []*valueobjects.BatchUpdateProject 批量更新任务清单值对象列表
// @return error 验证失败返回错误，否则返回 nil
func BatchUpdateProjectReqToValueObjects(
	req *dto.BatchUpdateProjectReq,
) ([]*valueobjects.BatchUpdateProject, error) {
	batchVOs := make([]*valueobjects.BatchUpdateProject, 0, len(req.Projects))
	for _, project := range req.Projects {
		id, err := idutil.ParseID(project.Id)
		if err != nil {
			return nil, err
		}
		vo, err := valueobjects.NewBatchUpdateProject(
			id,
			project.Name,
			project.Description,
			project.SortId,
		)
		if err != nil {
			return nil, err
		}
		batchVOs = append(batchVOs, vo)
	}
	return batchVOs, nil
}

// ProjectPreferenceEntityToGetRes 任务清单偏好实体转换响应体
// @param projectPreferenceEntity 任务清单偏好实体
// @return *dto.GetProjectPreferenceRes 获取任务清单偏好响应体
func ProjectPreferenceEntityToGetRes(
	projectPreferenceEntity *entities.ProjectPreference,
) *dto.GetProjectPreferenceRes {
	var res dto.GetProjectPreferenceRes
	res.Id = idutil.FormatID(projectPreferenceEntity.Id)
	res.CreatedAt = projectPreferenceEntity.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = projectPreferenceEntity.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = projectPreferenceEntity.DeletedAt.ToString(time.RFC3339)
	res.ProjectId = idutil.FormatID(projectPreferenceEntity.ProjectId)
	res.ViewType = projectPreferenceEntity.ViewType
	res.GetOptions = projectPreferenceEntity.GetOptions
	res.Columns = projectPreferenceEntity.Columns
	return &res
}

// UpdateProjectPreferenceReqToValueObject 更新任务清单偏好请求体转换值对象
// @param updateProjectPreferenceReq 更新任务清单偏好请求体
// @return *valueobjects.SaveProjectPreference 更新任务清单偏好值对象
// @return error 验证失败返回错误，否则返回 nil
func UpdateProjectPreferenceReqToValueObject(
	updateProjectPreferenceReq *dto.UpdateProjectPreferenceReq,
) (*valueobjects.SaveProjectPreference, error) {
	saveProjectPreferenceValueObject, err := valueobjects.NewSaveProjectPreference(
		domaintypes.ViewType(updateProjectPreferenceReq.ViewType),
		updateProjectPreferenceReq.GetOptions,
		updateProjectPreferenceReq.Columns,
	)
	if err != nil {
		return nil, err
	}
	return saveProjectPreferenceValueObject, nil
}
