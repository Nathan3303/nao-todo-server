package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PomodoroRepoImpl 实现 Pomodoro 仓库接口
type PomodoroRepoImpl struct {
	db *gorm.DB
}

// NewPomodoroRepo 创建 Pomodoro 仓库实例
func NewPomodoroRepo(db *gorm.DB) repositories.Pomodoro {
	return &PomodoroRepoImpl{db: db}
}

// Create 创建常用番茄工作
func (r *PomodoroRepoImpl) Create(
	ctx context.Context,
	vo *valueobjects.CreatePomodoro,
) (*entities.Pomodoro, error) {
	m := CreatePomodoroVOToModel(vo)
	tx := r.db.WithContext(ctx).Create(m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return PomodoroModel2Entity(m), nil
}

// GetById 获取常用番茄工作
func (r *PomodoroRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.Pomodoro, error) {
	var m models.Pomodoro
	tx := r.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userId).
		First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return PomodoroModel2Entity(&m), nil
}

// Update 更新常用番茄工作（PATCH 语义）
func (r *PomodoroRepoImpl) Update(
	ctx context.Context,
	userId int64,
	id int64,
	vo *valueobjects.UpdatePomodoro,
) (*entities.Pomodoro, error) {
	updateMap := UpdatePomodoroVOToMap(vo)
	tx := r.db.WithContext(ctx).
		Model(&models.Pomodoro{}).
		Where("id = ? AND user_id = ?", id, userId).
		Updates(updateMap)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return r.GetById(ctx, userId, id)
}

// Delete 删除常用番茄工作（软删除）
func (r *PomodoroRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	tx := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Delete(&models.Pomodoro{}, id)
	return tx.Error
}

// Archive 归档常用番茄工作
func (r *PomodoroRepoImpl) Archive(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	tx := r.db.WithContext(ctx).
		Model(&models.Pomodoro{}).
		Where("id = ? AND user_id = ?", id, userId).
		Update("archived_at", time.Now())
	return tx.Error
}

// Unarchive 取消归档常用番茄工作
func (r *PomodoroRepoImpl) Unarchive(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	tx := r.db.WithContext(ctx).
		Model(&models.Pomodoro{}).
		Where("id = ? AND user_id = ?", id, userId).
		Update("archived_at", nil)
	return tx.Error
}

// List 获取常用番茄工作列表
func (r *PomodoroRepoImpl) List(
	ctx context.Context,
	userId int64,
	pomodoroType uint8,
	name string,
	isArchived bool,
	page int,
	limit int,
	sort string,
) ([]*entities.Pomodoro, int64, error) {
	tx := r.db.WithContext(ctx).Model(&models.Pomodoro{}).
		Where("user_id = ?", userId)

	if pomodoroType > 0 {
		tx = tx.Where("type = ?", pomodoroType)
	}
	if name != "" {
		tx = tx.Where("name LIKE ?", "%"+name+"%")
	}
	if isArchived {
		tx = tx.Where("archived_at IS NOT NULL")
	} else {
		tx = tx.Where("archived_at IS NULL")
	}

	var total int64
	tx.Count(&total)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	if sort != "" {
		parts := strings.Split(sort, ":")
		if len(parts) == 2 {
			tx = tx.Order(utils.ToSnakeCase(parts[0]) + " " + parts[1])
		}
	}

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	offset := (page - 1) * limit

	var modelsList []*models.Pomodoro
	tx = tx.Offset(offset).Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	return PomodoroModels2Entities(modelsList), total, nil
}
