package project

import (
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
)

// 创建任务清单请求体转换值对象
// @param userId 用户 ID
// @param req 创建任务清单请求体
// @return *valueobjects.CreateProject 创建任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func CreateProjectReqToValueObject(
	userId int64,
	req *types.CreateProjectReq,
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

// 任务清单实体转换响应体
// @param projectEntity 任务清单实体
// @return *types.CreateProjectRes 创建任务清单响应体
func ProjectEntityToCreateRes(projectEntity *entities.Project) *types.CreateProjectRes {
	return &types.CreateProjectRes{
		Id:          strconv.FormatInt(projectEntity.Id, 10),
		UserId:      strconv.FormatInt(projectEntity.UserId, 10),
		Name:        projectEntity.Name,
		Description: projectEntity.Description,
		ArchivedAt:  projectEntity.ArchivedAt,
		CreatedAt:   projectEntity.CreatedAt,
		UpdatedAt:   projectEntity.UpdatedAt,
		DeactivedAt: projectEntity.DeactivedAt,
		SortId:      projectEntity.SortId,
	}
}

// 任务清单实体转换响应体
// @param projectEntity 任务清单实体
// @return *types.GetProjectRes 获取任务清单响应体
func ProjectEntityToGetRes(projectEntity *entities.Project) *types.GetProjectRes {
	return &types.GetProjectRes{
		Id:          strconv.FormatInt(projectEntity.Id, 10),
		UserId:      strconv.FormatInt(projectEntity.UserId, 10),
		Name:        projectEntity.Name,
		Description: projectEntity.Description,
		ArchivedAt:  projectEntity.ArchivedAt,
		CreatedAt:   projectEntity.CreatedAt,
		UpdatedAt:   projectEntity.UpdatedAt,
		DeactivedAt: projectEntity.DeactivedAt,
		SortId:      projectEntity.SortId,
	}
}

// 更新任务清单请求体转换值对象
// @param updateProjectReq 更新任务清单请求体
// @return *valueobjects.UpdateProject 更新任务清单值对象
// @return error 验证失败返回错误，否则返回 nil
func UpdateProjectReqToValueObject(
	updateProjectReq *types.UpdateProjectReq,
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

// 任务清单实体列表转换响应体列表
// @param projectEntities 任务清单实体列表
// @return []*types.GetProjectRes 任务清单响应体列表
func EntitiesToGetResList(projectEntities []*entities.Project) []*types.GetProjectRes {
	getResList := make([]*types.GetProjectRes, 0, len(projectEntities))
	for _, projectEntity := range projectEntities {
		getResList = append(getResList, ProjectEntityToGetRes(projectEntity))
	}
	return getResList
}

// 批量更新任务清单请求体转换值对象
// @param req 批量更新任务清单请求体
// @return []*valueobjects.BatchUpdateProject 批量更新任务清单值对象列表
// @return error 验证失败返回错误，否则返回 nil
func BatchUpdateProjectReqToValueObjects(
	req *types.BatchUpdateProjectReq,
) ([]*valueobjects.BatchUpdateProject, error) {
	batchVOs := make([]*valueobjects.BatchUpdateProject, 0, len(req.Projects))
	for _, project := range req.Projects {
		id, err := strconv.ParseInt(project.Id, 10, 64)
		if err != nil {
			return nil, err
		}
		vo, err := valueobjects.NewBatchUpdateProject(id, project.Name, project.Description, project.SortId)
		if err != nil {
			return nil, err
		}
		batchVOs = append(batchVOs, vo)
	}
	return batchVOs, nil
}

// 任务清单偏好实体转换响应体
// @param projectPreferenceEntity 任务清单偏好实体
// @return *types.GetProjectPreferenceRes 获取任务清单偏好响应体
func ProjectPreferenceEntityToGetRes(
	projectPreferenceEntity *entities.ProjectPreference,
) *types.GetProjectPreferenceRes {
	return &types.GetProjectPreferenceRes{
		Id:         strconv.FormatInt(projectPreferenceEntity.Id, 10),
		UserId:     strconv.FormatInt(projectPreferenceEntity.UserId, 10),
		ProjectId:  strconv.FormatInt(projectPreferenceEntity.ProjectId, 10),
		ViewType:   projectPreferenceEntity.ViewType,
		GetOptions: projectPreferenceEntity.GetOptions,
		Columns:    projectPreferenceEntity.Columns,
		CreatedAt:  projectPreferenceEntity.CreatedAt,
		UpdatedAt:  projectPreferenceEntity.UpdatedAt,
	}
}

// 更新任务清单偏好请求体转换值对象
// @param updateProjectPreferenceReq 更新任务清单偏好请求体
// @return *valueobjects.SaveProjectPreference 更新任务清单偏好值对象
// @return error 验证失败返回错误，否则返回 nil
func UpdateProjectPreferenceReqToValueObject(
	updateProjectPreferenceReq *types.UpdateProjectPreferenceReq,
) (*valueobjects.SaveProjectPreference, error) {
	saveProjectPreferenceValueObject, err := valueobjects.NewSaveProjectPreference(
		updateProjectPreferenceReq.ViewType,
		updateProjectPreferenceReq.GetOptions,
		updateProjectPreferenceReq.Columns,
	)
	if err != nil {
		return nil, err
	}
	return saveProjectPreferenceValueObject, nil
}
