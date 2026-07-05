package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"
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
	q *valueobjects.QueryPomodoro,
) ([]*entities.Pomodoro, int64, error) {
	tx := r.db.WithContext(ctx).Model(&models.Pomodoro{}).
		Where("user_id = ?", userId).
		Scopes(
			ByPomodoroType(q.Type),
			ByPomodoroName(q.Name),
			ByPomodoroArchivedFlag(q.IsArchived),
			query.Sort(q.Sort),
		)

	var total int64
	tx.Count(&total)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Limit <= 0 {
		q.Limit = 10
	}

	var modelsList []*models.Pomodoro
	tx = tx.Scopes(query.Paginate(q.Page, q.Limit)).Find(&modelsList)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	return PomodoroModels2Entities(modelsList), total, nil
}
