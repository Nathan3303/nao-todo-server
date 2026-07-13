package tag

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
)

type TagRepositoryImpl struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewTagRepo(db *gorm.DB, c *cache.Cache) repositories.TagRepository {
	return &TagRepositoryImpl{db: db, cache: c}
}

// GetById 获取标签信息
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return entities.Tag 标签实体
// @return error 错误
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

// Create 创建标签
// @param ctx 上下文
// @param userId 用户ID
// @param createTagValueObject 创建标签值对象
// @return entities.Tag 创建标签实体
// @return error 错误
func (tagRepo *TagRepositoryImpl) Create(
	ctx context.Context,
	userId int64,
	createTagValueObject *valueobjects.CreateTag,
) (*entities.Tag, error) {
	// 1. 转换标签实体为模型
	createValue := CreateTagValueObjectToModel(userId, createTagValueObject)
	// 2. 插入数据库
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).Create(createValue)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 返回结果
	entity := TagModel2Entity(createValue)
	// 4. 失效标签列表缓存
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))
	return entity, nil
}

// Update 更新标签
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @param updateTagValueObject 更新标签值对象
// @return error 错误
func (tagRepo *TagRepositoryImpl) Update(
	ctx context.Context,
	userId int64,
	tagId int64,
	updateTagValueObject *valueobjects.UpdateTag,
) error {
	// 1. 转换实体为模型
	var whereCond models.Tag
	whereCond.UserId = userId
	whereCond.ID = tagId
	updateCond := UpdateTagValueObjectToMap(updateTagValueObject)
	// 2. 更新数据
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where(&whereCond).
		Updates(updateCond)
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效标签列表缓存
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))
	return nil
}

// Delete 删除标签
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return error 错误
func (tagRepo *TagRepositoryImpl) Delete(ctx context.Context, userId int64, tagId int64) error {
	// 1. 转换实体为模型
	var whereCond models.Tag
	whereCond.UserId = userId
	whereCond.ID = tagId
	// 2. 删除
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where(&whereCond).
		Delete(&models.Tag{})
	if tx.Error != nil {
		return tx.Error
	}
	// 3. 失效标签列表缓存
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))
	return nil
}

// Get 获取所有标签
// @param ctx 上下文
// @param userId 用户ID
// @return []*entities.Tag 标签实体列表
// @return error 错误
func (tagRepo *TagRepositoryImpl) Get(ctx context.Context, userId int64) ([]*entities.Tag, error) {
	// 1. 优先读取缓存
	key := cache.TagListKey(userId)
	var cached []*entities.Tag
	if tagRepo.cache.Get(ctx, key, &cached) {
		return cached, nil
	}
	// 2. 转换实体为模型
	var findCond models.Tag
	findCond.UserId = userId
	// 3. 查询
	tagModelList := []*models.Tag{}
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Preload("Preference").
		Where(&findCond).
		Order("sort_id ASC").
		Find(&tagModelList)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 4. 转换为实体并写入缓存
	es := TagModelList2EntityList(tagModelList)
	tagRepo.cache.Set(ctx, key, es, time.Minute*30)
	// 5. 返回结果
	return es, nil
}

// GetMaxSortId 获取最大排序 ID
// @param ctx 上下文
// @param userId 用户ID
// @return maxSortId 最大排序 ID
func (tagRepo *TagRepositoryImpl) GetMaxSortId(
	ctx context.Context,
	userId int64,
) uint16 {
	var maxSortId uint16 = 255
	tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where("user_id = ?", userId).
		Pluck("MAX(sort_id)", &maxSortId)
	return maxSortId
}

// BatchUpdate 批量更新标签
func (tagRepo *TagRepositoryImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateTags []*valueobjects.BatchUpdateTag,
) ([]*entities.Tag, error) {
	tx := tagRepo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	updatedIds := make([]int64, 0, len(batchUpdateTags))

	for _, batchTag := range batchUpdateTags {
		whereCond := &models.Tag{}
		whereCond.ID = batchTag.Id
		whereCond.UserId = userId
		updateCond := BatchUpdateTagValueObjectToMap(batchTag)

		if err := tx.Model(&models.Tag{}).
			Where(whereCond).
			Updates(updateCond).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		updatedIds = append(updatedIds, batchTag.Id)
	}

	var updatedTags []*models.Tag
	if err := tx.Model(&models.Tag{}).
		Preload("Preference").
		Where("id IN ? AND user_id = ?", updatedIds, userId).
		Find(&updatedTags).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 失效标签列表缓存
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))

	return TagModelList2EntityList(updatedTags), nil
}
