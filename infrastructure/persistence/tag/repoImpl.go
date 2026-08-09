package tag

import (
	"context"
	"errors"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"
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

// Upsert 幂等写入标签：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）
func (tagRepo *TagRepositoryImpl) Upsert(
	ctx context.Context,
	userId int64,
	createTagValueObject *valueobjects.CreateTag,
) (*entities.Tag, bool, error) {
	if createTagValueObject.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		createTagValueObject.UpdatedAt = time.Now()
		entity, err := tagRepo.Create(ctx, userId, createTagValueObject)
		return entity, true, err
	}
	var existing models.Tag
	err := tagRepo.db.WithContext(ctx).Unscoped().
		Preload("Preference").
		Where("id = ? AND user_id = ?", createTagValueObject.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		createTagValueObject.UpdatedAt = time.Now()
		entity, createErr := tagRepo.Create(ctx, userId, createTagValueObject)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existing.UpdatedAt,
		createTagValueObject.CreatedAt, createTagValueObject.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return TagModel2Entity(&existing), false, nil
	}
	updateMap := CreateTagVOToUpdateMap(createTagValueObject)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	if err := tagRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Tag{}).
		Where("id = ? AND user_id = ?", createTagValueObject.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))
	var updated models.Tag
	if err := tagRepo.db.WithContext(ctx).Unscoped().
		Preload("Preference").
		Where("id = ? AND user_id = ?", createTagValueObject.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return TagModel2Entity(&updated), false, nil
}

// ListSync 增量同步标签列表：包含软删墓碑，(updated_at, id) keyset 游标 + 稳定排序 + limit（绕过缓存）
func (tagRepo *TagRepositoryImpl) ListSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.Tag, error) {
	tx := tagRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Tag{}).
		Where("user_id = ?", userId).
		Scopes(
			query.ByKeysetCursor(cursor, cursorID),
			query.SyncOrder(),
		)

	if limit <= 0 {
		limit = 100
	}

	var modelsList []*models.Tag
	tx = tx.Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return TagModelList2EntityList(modelsList), nil
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
	// 2. LWW 乐观锁：请求 updatedAt 早于库中版本时不更新
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where(&whereCond)
	if !updateTagValueObject.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", updateTagValueObject.UpdatedAt)
	}
	// 3. 更新数据
	tx = tx.Updates(updateCond)
	if tx.Error != nil {
		return tx.Error
	}
	// 4. 失效标签列表缓存
	tagRepo.cache.Del(ctx, cache.TagListKey(userId))
	return nil
}

// Delete 删除标签
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return error 错误
func (tagRepo *TagRepositoryImpl) Delete(ctx context.Context, userId int64, tagId int64) error {
	// 1. 软删同时推进 updated_at，保证删除墓碑可被增量拉取发现
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Where("id = ? AND user_id = ?", tagId, userId).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
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

// GetByIds 根据标签ID列表获取标签列表
// @param ctx 上下文
// @param userId 用户ID
// @param tagIds 标签ID列表
// @return []*entities.Tag 标签实体列表
// @return error 错误
func (tagRepo *TagRepositoryImpl) GetByIds(
	ctx context.Context,
	userId int64,
	tagIds []int64,
) ([]*entities.Tag, error) {
	// 1. 转换实体为模型
	var findCond models.Tag
	findCond.UserId = userId
	// 2. 结果切片
	var tags []*models.Tag
	tx := tagRepo.db.WithContext(ctx).Model(&models.Tag{}).
		Preload("Preference").
		Where(&findCond).
		Find(&tags, tagIds)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 2. 转换为实体
	es := TagModelList2EntityList(tags)
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
