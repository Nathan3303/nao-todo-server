package tag

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
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

// Get 获取标签偏好
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return entities.TagPreference 标签偏好实体
// @return error 错误
func (t *TagPreferenceRepoImpl) Get(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*entities.TagPreference, error) {
	// 1. valueobject 转 model
	var whereCond models.TagPreference
	whereCond.UserId = userId
	whereCond.TagId = tagId
	// 2. 查询
	var preference models.TagPreference
	tx := t.db.WithContext(ctx).Model(&models.TagPreference{}).
		Where(&whereCond).
		First(&preference)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 返回
	return TagPreferenceModel2Entity(&preference), nil
}

// Save 保存标签偏好
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @param saveTagPreferenceValueObject 保存标签偏好值对象
// @return error 错误
func (t *TagPreferenceRepoImpl) Save(
	ctx context.Context,
	userId int64,
	tagId int64,
	saveTagPreferenceValueObject *valueobjects.SaveTagPreference,
) error {
	// 1. valueobject 转 model
	preference := UpdateTagPreferenceValueObjectToModel(
		saveTagPreferenceValueObject,
	)
	// 2. 定义查找模型
	var whereCond models.TagPreference
	whereCond.UserId = userId
	whereCond.TagId = tagId
	// 2. 查找原有记录
	var existPreference models.TagPreference
	findTx := t.db.WithContext(ctx).
		Model(&models.TagPreference{}).
		Where(&whereCond).
		First(&existPreference)
		// 3. 检查是否存在并执行对应动作
	var saveTx *gorm.DB
	if findTx.Error != gorm.ErrRecordNotFound || existPreference.ID > 0 {
		// 更新
		saveTx = t.db.WithContext(ctx).
			Model(&models.TagPreference{}).
			Where(&whereCond).
			Updates(preference)
	} else {
		// 新增
		saveTx = t.db.WithContext(ctx).
			Model(&models.TagPreference{}).
			Where(&whereCond).
			Create(preference)
	}
	if saveTx.Error != nil {
		return saveTx.Error
	}
	// 3. model 转 valueobject
	return nil
}
