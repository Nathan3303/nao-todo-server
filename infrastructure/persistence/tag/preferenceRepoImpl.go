package tag

import (
	"context"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/vo"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type TagPreferenceRepoImpl struct {
	db *gorm.DB
}

func NewTagPreferenceRepo(db *gorm.DB) repositories.TagPreference {
	return &TagPreferenceRepoImpl{
		db: db,
	}
}

// Get implements [repositories.TagPreference].
func (t *TagPreferenceRepoImpl) Get(
	ctx context.Context,
	whereVO *vo.TagPreference,
) (*vo.TagPreference, error) {
	// 1. valueobject 转 model
	whereCond := TagPreferenceVO2Model(whereVO)
	// 2. 查询
	var preference models.TagPreference
	tx := t.db.WithContext(ctx).Model(&models.TagPreference{}).
		Where(&whereCond).
		First(&preference)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 返回
	return TagPreferenceModel2VO(&preference), nil
}

// Save implements [repositories.TagPreference].
func (t *TagPreferenceRepoImpl) Save(
	ctx context.Context,
	preferenceVO *vo.TagPreference,
) (*vo.TagPreference, error) {
	// 1. valueobject 转 model
	preference := TagPreferenceVO2Model(preferenceVO)
	// 2. 查找原有记录
	var existPreference models.TagPreference
	findTx := t.db.WithContext(ctx).
		Model(&models.TagPreference{}).
		Where("user_id = ? AND tag_id = ?", preferenceVO.UserId, preferenceVO.TagId).
		First(&existPreference)
		// 3. 检查是否存在并执行对应动作
	var saveTx *gorm.DB
	if findTx.Error != gorm.ErrRecordNotFound || existPreference.ID > 0 {
		// 更新
		saveTx = t.db.WithContext(ctx).
			Model(&models.TagPreference{}).
			Where("user_id = ? AND tag_id = ?", preferenceVO.UserId, preferenceVO.TagId).
			Updates(preference)
	} else {
		// 新增
		saveTx = t.db.WithContext(ctx).
			Model(&models.TagPreference{}).
			Where("user_id = ? AND tag_id = ?", preferenceVO.UserId, preferenceVO.TagId).
			Create(preference)
	}
	if saveTx.Error != nil {
		return nil, saveTx.Error
	}
	// 3. model 转 valueobject
	return TagPreferenceModel2VO(preference), nil
}
