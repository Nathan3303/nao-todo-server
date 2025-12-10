package tag

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type TagRepositoryImpl struct {
	db *gorm.DB
}

func NewTagRepo(db *gorm.DB) repositories.TagRepository {
	return &TagRepositoryImpl{db: db}
}

/*
 * Get tag by id
 * 根据标签ID获取标签信息
 */
func (tagRepo *TagRepositoryImpl) GetById(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*entities.Tag, error) {
	// 1. 创建查找模型
	findCond := &models.Tag{}
	findCond.UserId = userId
	findCond.ID = tagId
	// 2. 创建结果模型
	tagModel := &models.Tag{}
	// 3. 查询
	tx := tagRepo.db.WithContext(ctx).
		Preload("Preference").
		Where(findCond).
		First(tagModel)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 返回结果
	return TagModel2Entity(tagModel), nil
}

/*
 * Create tag
 * 创建标签
 */
func (tagRepo *TagRepositoryImpl) Create(
	ctx context.Context,
	createEntity *entities.Tag,
) (*entities.Tag, error) {
	// 1. 转换标签实体为模型
	createValue := TagEntity2Model(createEntity)
	// 2. 插入数据库
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).Create(createValue)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 返回结果
	return TagModel2Entity(createValue), nil
}

/*
 * Update tag
 * 更新标签
 */
func (tagRepo *TagRepositoryImpl) Update(
	ctx context.Context,
	whereEntity *entities.Tag,
	updateEntity *entities.Tag,
) error {
	// 1. 转换实体为模型
	findCond := TagEntity2Model(whereEntity)
	updateCond := TagEntity2Model(updateEntity)
	// 2. 更新数据
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where(findCond).
		Updates(updateCond)
	return tx.Error
}

/*
 * Delete tag
 * 删除标签
 */
func (tagRepo *TagRepositoryImpl) Delete(
	ctx context.Context,
	whereEntity *entities.Tag,
) error {
	// 1. 转换实体为模型
	findCond := TagEntity2Model(whereEntity)
	// 2. 删除
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where(findCond).
		Delete(&models.Tag{})
	return tx.Error
}

/*
 * Get tag
 * 获取所有标签
 */
func (tagRepo *TagRepositoryImpl) Get(
	ctx context.Context,
	whereEntity *entities.Tag,
) ([]*entities.Tag, error) {
	// 1. 转换实体为模型
	findCond := TagEntity2Model(whereEntity)
	// 2. 查询
	tagModelList := []*models.Tag{}
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Preload("Preference").
		Where(findCond).
		Find(&tagModelList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 返回结果
	return TagModelList2EntityList(tagModelList), nil
}
