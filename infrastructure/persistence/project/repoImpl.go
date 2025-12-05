package project

import (
	"context"
	"errors"
	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
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

/*
 * Create 创建清单
 */
func (projectRepo *ProjectRepoImpl) Create(
	ctx context.Context,
	project *entities.Project,
) (*entities.Project, error) {
	// 1. 转换为模型
	m := Entity2Model(project)
	// 2. 入库
	tx := projectRepo.db.WithContext(ctx).Create(m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	// 3. 转换为实体并返回结果
	return Model2Entity(m), nil
}

/*
 * GetById 获取单个清单详情
 */
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

/*
 * Update 更新清单
 */
func (projectRepo *ProjectRepoImpl) Update(
	ctx context.Context,
	whereEntity *entities.Project,
	updateEntity *entities.Project,
) error {
	// 1. 转换为模型
	whereCond := Entity2Model(whereEntity)
	updateCond := Entity2Model(updateEntity)
	// 2. 更新数据库
	tx := projectRepo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(whereCond).
		Updates(updateCond)
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

/*
 * Delete 删除清单
 */
func (projectRepo *ProjectRepoImpl) Delete(
	ctx context.Context,
	whereEntity *entities.Project,
) error {
	// 1. 转换为模型
	whereCond := Entity2Model(whereEntity)
	// 2. 删除数据库
	tx := projectRepo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where(whereCond).
		Delete(&models.Project{})
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	return nil
}

/*
 * Restore 恢复清单
 */
func (projectRepo *ProjectRepoImpl) Restore(
	ctx context.Context,
	whereEntity *entities.Project,
) error {
	// 1. 转换为模型
	whereCond := Entity2Model(whereEntity)
	// 2. 恢复数据库
	tx := projectRepo.db.WithContext(ctx).Unscoped().Model(&models.Project{}).
		Where(whereCond).
		Update("deleted_at", nil)
	// 3. 返回结果
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("清单不存在")
	}
	return nil
}

/*
 * Archive 归档清单
 */
func (projectRepo *ProjectRepoImpl) Archive(
	ctx context.Context,
	whereEntity *entities.Project,
) error {
	// 1. 转换为模型
	whereCond := Entity2Model(whereEntity)
	// 2. 归档数据库
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where(whereCond).
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

/*
 * Unarchive 取消归档清单
 */
func (projectRepo *ProjectRepoImpl) Unarchive(
	ctx context.Context,
	whereEntity *entities.Project,
) error {
	// 1. 转换为模型
	whereCond := Entity2Model(whereEntity)
	// 2. 取消归档数据库
	tx := projectRepo.db.WithContext(ctx).Model(&models.Project{}).
		Where(whereCond).
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

/*
 * GetByUserId 获取用户所有清单
 */
func (projectRepo *ProjectRepoImpl) GetByUserId(
	ctx context.Context,
	userId int64,
) ([]*entities.Project, error) {
	// 1. 从数据库中查询
	var ms []*models.Project
	tx := projectRepo.db.WithContext(ctx).
		Preload("Preference").
		Where("user_id = ?", userId).
		Find(&ms)
	// 2. 转换为实体并返回结果
	if tx.Error != nil {
		return nil, tx.Error
	}
	return Models2Entities(ms), nil
}
