package project

import (
	"context"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/vo"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type ProjectPreferenceRepoImpl struct {
	db *gorm.DB
}

func NewProjectPreferenceRepo(db *gorm.DB) repositories.ProjectPreference {
	return &ProjectPreferenceRepoImpl{db: db}
}

// Get implements [repositories.ProjectPreference].
func (projectPreferenceRepo *ProjectPreferenceRepoImpl) Get(
	ctx context.Context,
	preferenceVO vo.ProjectPreference,
) (*vo.ProjectPreference, error) {
	// 1. valueobject 转 model
	whereCond := PreferenceVO2Model(&preferenceVO)
	// 2. 查询
	var preference models.ProjectPreference
	tx := projectPreferenceRepo.db.WithContext(ctx).Model(&models.ProjectPreference{}).
		Where(&whereCond).
		First(&preference)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. model 转 valueobject
	return PreferenceModel2VO(&preference), nil
}

// Save implements [repositories.ProjectPreference].
func (projectPreferenceRepo *ProjectPreferenceRepoImpl) Save(
	ctx context.Context,
	preferenceVO *vo.ProjectPreference,
) (*vo.ProjectPreference, error) {
	panic("unimplemented")
}
