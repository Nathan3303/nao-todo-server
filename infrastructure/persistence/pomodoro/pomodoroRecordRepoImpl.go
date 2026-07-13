package pomodoro

import (
	"context"
	"naotodoserver/domain/pomodoro/entities"
	"naotodoserver/domain/pomodoro/repositories"
	"naotodoserver/domain/pomodoro/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
	query "naotodoserver/infrastructure/utils/query"

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
