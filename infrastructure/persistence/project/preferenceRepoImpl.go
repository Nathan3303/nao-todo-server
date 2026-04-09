package project

import (
	"context"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type ProjectPreferenceRepoImpl struct {
	db *gorm.DB
}

func NewProjectPreferenceRepo(db *gorm.DB) repositories.ProjectPreference {
	return &ProjectPreferenceRepoImpl{db: db}
}

// Get 获取项目偏好设置
// @param ctx 上下文
// @param preferenceVO 项目偏好设置
// @return 项目偏好设置
// @return error 错误
func (projectPreferenceRepo *ProjectPreferenceRepoImpl) Get(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.ProjectPreference, error) {
	// 1. valueobject 转 model
	var whereCond models.ProjectPreference
	whereCond.UserId = userId
	whereCond.ProjectId = projectId
	// 2. 查询
	var preference models.ProjectPreference
	tx := projectPreferenceRepo.db.WithContext(ctx).Model(&models.ProjectPreference{}).
		Where(&whereCond).
		First(&preference)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. model 转 valueobject
	return PreferenceModel2Entity(&preference), nil
}

// Save 保存项目偏好设置
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @param saveProjectPreference 项目偏好设置
// @return 项目偏好设置
// @return error 错误
func (projectPreferenceRepo *ProjectPreferenceRepoImpl) Save(
	ctx context.Context,
	userId int64,
	projectId int64,
	saveProjectPreference *valueobjects.SaveProjectPreference,
) error {
	// 1. valueobject 转 model
	var whereCond models.ProjectPreference
	whereCond.UserId = userId
	whereCond.ProjectId = projectId
	// 2. 保存
	var existPreference models.ProjectPreference
	tx := projectPreferenceRepo.db.WithContext(ctx).
		Model(&models.ProjectPreference{}).
		Where(&whereCond).
		First(&existPreference)
	if tx.Error == gorm.ErrRecordNotFound {
		// 转换为 model
		preferenceModel := UpdateProjectPrefrenceValueObjectToModel(
			userId,
			projectId,
			saveProjectPreference,
		)
		// 新增
		projectPreferenceRepo.db.WithContext(ctx).
			Model(&models.ProjectPreference{}).Create(preferenceModel)
	} else {
		// 更新
		tx.Updates(models.ProjectPreference{
			ViewType:   saveProjectPreference.ViewType,
			GetOptions: saveProjectPreference.GetOptions,
			Columns:    saveProjectPreference.Columns,
		})
	}
	// 3. 返回结果
	return nil
}
