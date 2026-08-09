package project

import (
	"context"
	"errors"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"
	"time"

	"gorm.io/gorm"
)

type ProjectRepoImpl struct {
	db    *gorm.DB
	cache *cache.Cache
}

func NewProjectRepo(db *gorm.DB, c *cache.Cache) repositories.Project {
	return &ProjectRepoImpl{db: db, cache: c}
}

// Create 创建清单
// @param ctx 上下文
// @param createProjectValueObject 项目创建值对象
// @return 项目
// @return error 错误
func (projectRepo *ProjectRepoImpl) Create(
	ctx context.Context,
	createProjectValueObject *valueobjects.CreateProject,
) (*entities.Project, error) {
	// 1. 转换为模型
	m := CreateProjectValueObject2Model(createProjectValueObject)
	// 2. 入库
	tx := projectRepo.db.WithContext(ctx).Create(m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 转换为实体并返回结果
	entity := Model2Entity(m)
	// 4. 失效项目列表缓存
	projectRepo.cache.Del(ctx, cache.ProjectListKey(createProjectValueObject.UserId))
	return entity, nil
}

// Upsert 幂等写入任务清单：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）
func (projectRepo *ProjectRepoImpl) Upsert(
	ctx context.Context,
	userId int64,
	createProjectValueObject *valueobjects.CreateProject,
) (*entities.Project, bool, error) {
	if createProjectValueObject.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		createProjectValueObject.UpdatedAt = time.Now()
		entity, err := projectRepo.Create(ctx, createProjectValueObject)
		return entity, true, err
	}
	var existing models.Project
	err := projectRepo.db.WithContext(ctx).Unscoped().
		Preload("Preference").
		Where("id = ? AND user_id = ?", createProjectValueObject.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		createProjectValueObject.UpdatedAt = time.Now()
		entity, createErr := projectRepo.Create(ctx, createProjectValueObject)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existing.UpdatedAt,
		createProjectValueObject.CreatedAt, createProjectValueObject.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return Model2Entity(&existing), false, nil
	}
	updateMap := CreateProjectVOToUpdateMap(createProjectValueObject)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	if err := projectRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Project{}).
		Where("id = ? AND user_id = ?", createProjectValueObject.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	projectRepo.cache.Del(ctx, cache.ProjectListKey(userId))
	var updated models.Project
	if err := projectRepo.db.WithContext(ctx).Unscoped().
		Preload("Preference").
		Where("id = ? AND user_id = ?", createProjectValueObject.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return Model2Entity(&updated), false, nil
}

// GetById 获取单个清单详情
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return 项目
// @return error 错误
func (projectRepo *ProjectRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	projectId int64,
) (*entities.Project, error) {
	// 1. 从数据库中查询
	var m models.Project
	tx := projectRepo.db.WithContext(ctx).
		Preload("Preference").
		Where("user_id = ? AND id = ?", userId, projectId).
		First(&m)
	// 2. 转换为实体并返回结果
	if tx.Error != nil {
		return nil, tx.Error
	}
	return Model2Entity(&m), nil
}

// Update 更新清单
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @param updateProjectValueObject 更新项目值对象
// @return error 错误
func (projectRepo *ProjectRepoImpl) Update(
	ctx context.Context,
	userId int64,
	projectId int64,
	updateProjectValueObject *valueobjects.UpdateProject,
) error {
	// 1. 转换为模型
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	// 2. 更新值对象转换为 map 格式
	updateCond := UpdateProjectValueObjectToMap(updateProjectValueObject)
	// 3. LWW 乐观锁：请求 updatedAt 早于库中版本时不更新
	tx := projectRepo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&whereCond)
	if !updateProjectValueObject.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", updateProjectValueObject.UpdatedAt)
	}
	// 4. 更新数据库
	tx = tx.Updates(updateCond)
	// 4. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	// 5. 失效项目列表缓存
	projectRepo.cache.Del(ctx, cache.ProjectListKey(userId))
	return nil
}

// UpdateState 更新清单状态（归档/停用时间字段）
// 根据实体状态写入归档时间/停用时间：
// - ShouldUpdate 时更新（IsSetToNull 时置空）
// - 字段为空（IsNull）时视为取消归档/恢复清单，置空对应列
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @param archivedAt 归档时间
// @param deactivedAt 停用时间
// @return error 错误
func (projectRepo *ProjectRepoImpl) UpdateState(
	ctx context.Context,
	userId int64,
	projectId int64,
	archivedAt types.NullableTime,
	deactivedAt types.NullableTime,
) error {
	// 1. 构建更新 map
	updateMap := make(map[string]any)
	if archivedAt.ShouldUpdate() {
		if archivedAt.IsSetToNull() {
			updateMap["archived_at"] = nil
		} else {
			updateMap["archived_at"] = archivedAt.ToSqlNullTime()
		}
	} else if archivedAt.IsNull {
		updateMap["archived_at"] = nil
	}
	if deactivedAt.ShouldUpdate() {
		if deactivedAt.IsSetToNull() {
			updateMap["deactived_at"] = nil
		} else {
			updateMap["deactived_at"] = deactivedAt.ToSqlNullTime()
		}
	} else if deactivedAt.IsNull {
		updateMap["deactived_at"] = nil
	}
	// 2. 更新数据库
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	tx := dbs.DBFrom(ctx, projectRepo.db).WithContext(ctx).
		Model(&models.Project{}).
		Where(&whereCond).
		Updates(updateMap)
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("清单不存在")
	}
	// 4. 失效项目列表缓存
	projectRepo.cache.Del(ctx, cache.ProjectListKey(userId))
	return nil
}

// BatchUpdate 批量更新清单
func (projectRepo *ProjectRepoImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateProjects []*valueobjects.BatchUpdateProject,
) ([]*entities.Project, error) {
	tx := projectRepo.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	updatedIds := make([]int64, 0, len(batchUpdateProjects))

	for _, batchProject := range batchUpdateProjects {
		whereCond := &models.Project{}
		whereCond.ID = batchProject.Id
		whereCond.UserId = userId
		updateCond := BatchUpdateProjectValueObjectToMap(batchProject)

		if err := tx.Model(&models.Project{}).
			Where(whereCond).
			Updates(updateCond).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		updatedIds = append(updatedIds, batchProject.Id)
	}

	var updatedProjects []*models.Project
	if err := tx.Model(&models.Project{}).
		Preload("Preference").
		Where("id IN ? AND user_id = ?", updatedIds, userId).
		Find(&updatedProjects).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	// 失效项目列表缓存
	projectRepo.cache.Del(ctx, cache.ProjectListKey(userId))

	return Models2Entities(updatedProjects), nil
}

// GetByUserId 获取用户所有清单
// @param ctx 上下文
// @param userId 用户ID
// @return 项目列表
// @return error 错误
func (projectRepo *ProjectRepoImpl) GetByUserId(
	ctx context.Context,
	userId int64,
) ([]*entities.Project, error) {
	// 1. 优先读取缓存
	key := cache.ProjectListKey(userId)
	var cached []*entities.Project
	if projectRepo.cache.Get(ctx, key, &cached) {
		return cached, nil
	}
	// 2. 从数据库中查询
	var ms []*models.Project
	tx := projectRepo.db.WithContext(ctx).
		Preload("Preference").
		Where("user_id = ?", userId).
		Order("sort_id ASC").
		Find(&ms)
	// 3. 转换为实体并返回结果
	if tx.Error != nil {
		return nil, tx.Error
	}
	es := Models2Entities(ms)
	// 4. 写入缓存
	projectRepo.cache.Set(ctx, key, es, time.Minute*30)
	return es, nil
}

// GetMaxSortId 获取最大排序 ID
// @param ctx 上下文
// @param userId 用户ID
// @return maxSortId 最大排序 ID
func (projectRepo *ProjectRepoImpl) GetMaxSortId(
	ctx context.Context,
	userId int64,
) uint16 {
	var maxSortId uint16 = 255
	projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where("user_id = ?", userId).
		Pluck("MAX(sort_id)", &maxSortId)
	return maxSortId
}

// DeleteDeactivatedProjects 删除已注销的任务清单
func (projectRepo *ProjectRepoImpl) DeleteDeactivatedProjects(
	ctx context.Context,
	dayOffset int8,
) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -1*int(dayOffset))
	tx := projectRepo.db.
		WithContext(ctx).
		Model(&models.Project{}).
		Where("deactived_at < ?", cutoff).
		Delete(&models.Project{})
	return tx.RowsAffected, tx.Error
}

// ListSync 增量同步任务清单列表：包含软删墓碑，(updated_at, id) keyset 游标 + 稳定排序 + limit（绕过缓存）
func (projectRepo *ProjectRepoImpl) ListSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.Project, error) {
	tx := projectRepo.db.WithContext(ctx).Unscoped().
		Model(&models.Project{}).
		Where("user_id = ?", userId).
		Scopes(
			query.ByKeysetCursor(cursor, cursorID),
			query.SyncOrder(),
		)

	if limit <= 0 {
		limit = 100
	}

	var modelsList []*models.Project
	tx = tx.Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return Models2Entities(modelsList), nil
}
