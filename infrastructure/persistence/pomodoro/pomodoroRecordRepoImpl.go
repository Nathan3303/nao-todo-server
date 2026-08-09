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

// PomodoroRecordRepoImpl 实现 PomodoroRecordRepository 接口
type PomodoroRecordRepoImpl struct {
	db *gorm.DB
}

// NewPomodoroRecordRepo 创建 PomodoroRecordRepository 实例
func NewPomodoroRecordRepo(db *gorm.DB) repositories.PomodoroRecord {
	return &PomodoroRecordRepoImpl{db: db}
}

// Create 创建 PomodoroRecord
// 若记录关联了常用番茄工作（PomodoroId > 0），在同一事务内原子累加其 TotalDuration，
// 保证记录落库与时长累计要么同时成功、要么同时回滚。
func (r *PomodoroRecordRepoImpl) Create(
	ctx context.Context,
	vo *valueobjects.CreatePomodoroRecord,
) (*entities.PomodoroRecord, error) {
	m := CreatePomodoroRecordVOToModel(vo)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(m).Error; err != nil {
			return err
		}
		if vo.PomodoroId > 0 {
			return tx.Model(&models.Pomodoro{}).
				Where("id = ? AND user_id = ?", vo.PomodoroId, vo.UserId).
				Update("total_duration", gorm.Expr("total_duration + ?", uint64(vo.Duration))).
				Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return PomodoroRecordModel2Entity(m), nil
}

// Upsert 幂等写入番茄工作记录：客户端指定 id 时创建或覆盖
// 语义与 Task.Upsert 一致（LWW + create 冲突检测）。
// 覆盖分支仅更新 Create VO 表达的字段，不触碰 archived_at，也不调整 totalDuration 累加。
func (r *PomodoroRecordRepoImpl) Upsert(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreatePomodoroRecord,
) (*entities.PomodoroRecord, bool, error) {
	if vo.Id == 0 {
		// 服务器时间为唯一基准：新建实体 updated_at 落服务器 now
		vo.UpdatedAt = time.Now()
		entity, err := r.Create(ctx, vo)
		return entity, true, err
	}
	var existing models.PomodoroRecord
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
		return PomodoroRecordModel2Entity(&existing), false, nil
	}
	updateMap := CreatePomodoroRecordVOToUpdateMap(vo)
	// 服务器时间为唯一基准：覆盖写入 updated_at 用服务器 now
	updateMap["updated_at"] = time.Now()
	if err := r.db.WithContext(ctx).Unscoped().
		Model(&models.PomodoroRecord{}).
		Where("id = ? AND user_id = ?", vo.Id, userId).
		UpdateColumns(updateMap).Error; err != nil {
		return nil, false, err
	}
	var updated models.PomodoroRecord
	if err := r.db.WithContext(ctx).Unscoped().
		Where("id = ? AND user_id = ?", vo.Id, userId).
		First(&updated).Error; err != nil {
		return nil, false, err
	}
	return PomodoroRecordModel2Entity(&updated), false, nil
}

// GetById 获取 PomodoroRecord
func (r *PomodoroRecordRepoImpl) GetById(
	ctx context.Context,
	userId int64,
	id int64,
) (*entities.PomodoroRecord, error) {
	var m models.PomodoroRecord
	tx := r.db.
		WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userId).
		First(&m)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return PomodoroRecordModel2Entity(&m), nil
}

// List 获取 PomodoroRecord 列表
func (r *PomodoroRecordRepoImpl) List(
	ctx context.Context,
	userId int64,
	q *valueobjects.QueryPomodoroRecord,
) ([]*entities.PomodoroRecord, int64, error) {
	tx := r.db.WithContext(ctx).Model(&models.PomodoroRecord{}).
		Where("user_id = ?", userId).
		Scopes(
			ByPomodoroRecordSessionId(q.SessionId),
			ByPomodoroRecordPomodoroId(q.PomodoroId),
			ByPomodoroRecordTimeRange(q.StartTime, q.EndTime),
			ByPomodoroRecordTaskId(q.TaskId),
			ByPomodoroRecordTaskName(q.TaskName),
			ByPomodoroType(uint8(q.Type)),
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

	var modelsList []*models.PomodoroRecord
	tx = tx.Scopes(query.Paginate(q.Page, q.Limit)).Find(&modelsList)
	if tx.Error != nil {
		return nil, 0, tx.Error
	}

	return PomodoroRecordModels2Entities(modelsList), total, nil
}

// ListSync 增量同步番茄工作记录列表：包含软删墓碑，(updated_at, id) keyset 游标 + 稳定排序 + limit
func (r *PomodoroRecordRepoImpl) ListSync(
	ctx context.Context,
	userId int64,
	cursor time.Time,
	cursorID int64,
	limit int,
) ([]*entities.PomodoroRecord, error) {
	tx := r.db.WithContext(ctx).Unscoped().
		Model(&models.PomodoroRecord{}).
		Where("user_id = ?", userId).
		Scopes(
			query.ByKeysetCursor(cursor, cursorID),
			query.SyncOrder(),
		)

	if limit <= 0 {
		limit = 100
	}

	var modelsList []*models.PomodoroRecord
	tx = tx.Limit(limit).Find(&modelsList)
	if tx.Error != nil {
		return nil, tx.Error
	}

	return PomodoroRecordModels2Entities(modelsList), nil
}
