package project

import (
	"context"
	"errors"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"time"

	"gorm.io/gorm"
)

type ProjectRepoImpl struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) repositories.Project {
	return &ProjectRepoImpl{db: db}
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
	return Model2Entity(m), nil
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
	// 3. 更新数据库
	tx := projectRepo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&whereCond).
		Updates(updateCond)
	// 4. 返回结果
	return tx.Error
}

// Delete 删除清单（软删除，设置 deactived_at）
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (projectRepo *ProjectRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	// 2. 更新 deactived_at 为当前时间
	tx := projectRepo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(&whereCond).
		Update("deactived_at", time.Now())
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

// Restore 恢复清单（取消软删除，设置 deactived_at 为 nil）
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (projectRepo *ProjectRepoImpl) Restore(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	// 2. 恢复 deactived_at 为 nil
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where(&whereCond).
		Update("deactived_at", nil)
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("清单不存在")
	}
	return nil
}

// Archive 归档清单
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (projectRepo *ProjectRepoImpl) Archive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	// 2. 归档数据库
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where(&whereCond).
		Update("archived_at", time.Now())
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("清单不存在")
	}
	return nil
}

// Unarchive 取消归档清单
// @param ctx 上下文
// @param userId 用户ID
// @param projectId 项目ID
// @return error 错误
func (projectRepo *ProjectRepoImpl) Unarchive(
	ctx context.Context,
	userId int64,
	projectId int64,
) error {
	// 1. 转换为模型
	var whereCond models.Project
	whereCond.UserId = userId
	whereCond.ID = projectId
	// 2. 取消归档数据库
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where(&whereCond).
		Update("archived_at", nil)
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("清单不存在")
	}
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
	// 1. 从数据库中查询
	var ms []*models.Project
	tx := projectRepo.db.WithContext(ctx).
		Preload("Preference").
		Where("user_id = ?", userId).
		Order("sort_id ASC").
		Find(&ms)
	// 2. 转换为实体并返回结果
	if tx.Error != nil {
		return nil, tx.Error
	}
	return Models2Entities(ms), nil
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
func (projectRepo *ProjectRepoImpl) DeleteDeactivatedProjects(ctx context.Context, dayOffset int8) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -1*int(dayOffset))
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where("deactived_at < ?", cutoff).
		Delete(&models.Project{})
	return tx.RowsAffected, tx.Error
}
