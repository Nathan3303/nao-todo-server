package pomodoro

import (
	"context"
	"errors"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/domain/types"
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

// Upsert 幂等写入常用番茄工作：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）
func (r *PomodoroRepoImpl) Upsert(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreatePomodoro,
) (*entities.Pomodoro, bool, error) {
	if vo.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, err := r.Create(ctx, vo)
		return entity, true, err
	}
	var existing models.Pomodoro
	err := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, createErr := r.Create(ctx, vo)
		return entity, true, createErr
	}
	if err != nil {
		return nil, false, err
	}
	outcome, err := types.DecideUpsert(
		existing.CreatedAt, existing.UpdatedAt,
		vo.CreatedAt, vo.UpdatedAt,
		time.Minute,
	)
	if err != nil {
		return nil, false, err
	}
	if outcome == types.UpsertNoop {
		return PomodoroModel2Entity(&existing), false, nil
	}
	updateMap := CreatePomodoroVOToUpdateMap(vo)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	// 覆盖已软删记录（墓碑）时复活：未携带删除时间时显式清 deleted_at
	if _, ok := updateMap["deleted_at"]; !ok {
		updateMap["deleted_at"] = gorm.Expr("NULL")
	}
	if err := r.db.WithContext(ctx).Unscoped().
		Model(&models.Pomodoro{}).
		Where("id = ? AND user_id = ?", vo.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	var updated models.Pomodoro
	if err := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return PomodoroModel2Entity(&updated), false, nil
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
		Where("id = ? AND user_id = ?", id, userId)
	// LWW 乐观锁：请求 updatedAt 早于库中版本时不更新
	if !vo.UpdatedAt.IsZero() {
		tx = tx.Where("updated_at <= ?", vo.UpdatedAt)
	}
	tx = tx.Updates(updateMap)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return r.GetById(ctx, userId, id)
}

// Delete 删除常用番茄工作（软删除，同时推进 updated_at 保证删除墓碑可增量发现）
func (r *PomodoroRepoImpl) Delete(
	ctx context.Context,
	userId int64,
	id int64,
) error {
	tx := r.db.WithContext(ctx).
		Model(&models.Pomodoro{}).
		Where("id = ? AND user_id = ?", id, userId).
		UpdateColumns(map[string]any{
			"deleted_at": time.Now(),
			"updated_at": time.Now(),
		})
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
			ByPomodoroType(uint8(q.Type)),
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

// ListSync 增量同步常用番茄工作列表：包含软删墓碑，(updated_at, id) keyset 游标 + 稳定排序 + limit
func (r *PomodoroRepoImpl) ListSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.Pomodoro, error) {
	tx := r.db.WithContext(ctx).Unscoped().
		Model(&models.Pomodoro{}).
		Where("user_id = ?", userId).
		Scopes(
			query.ByKeysetCursor(cursor, cursorID),
			query.SyncOrder(),
		)

	if limit <= 0 {
		limit = 100
	}

	var modelsList []*models.Pomodoro
	tx = tx.Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return PomodoroModels2Entities(modelsList), nil
}
