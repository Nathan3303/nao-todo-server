package service

import (
	"context"
	"database/sql"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
	"strconv"
	"strings"
	"time"
)

func NewTaskDomain(taskRepo repositories.Task) TaskDomain {
	return &TaskDomainImpl{taskRepo: taskRepo}
}

// GetById 获取单个任务信息
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return 任务实体
// @return error 错误信息
func (taskDomain *TaskDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	return taskDomain.taskRepo.GetById(ctx, userId, taskId)
}

// Create 创建任务
// @param ctx 上下文
// @param userId 用户ID
// @param createEntity 任务实体
// @return 任务实体
// @return error 错误信息
func (taskDomain *TaskDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createTaskValueObject *valueobjects.CreateTask,
) (*entities.Task, error) {
	return taskDomain.taskRepo.Create(ctx, userId, createTaskValueObject)
}

// Update 更新任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @param updateTaskValueObject 更新任务值对象
// @return error 错误
func (taskDomain *TaskDomainImpl) Update(
	ctx context.Context,
	userId int64,
	taskId int64,
	updateTaskValueObject *valueobjects.UpdateTask,
) error {
	return taskDomain.taskRepo.Update(ctx, userId, taskId, updateTaskValueObject)
}

// Delete 删除任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskDomain *TaskDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	return taskDomain.taskRepo.Delete(ctx, userId, taskId)
}

// Restore 恢复任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return error 错误
func (taskDomain *TaskDomainImpl) Restore(
	ctx context.Context,
	userId int64,
	taskId int64,
) error {
	return taskDomain.taskRepo.Restore(ctx, userId, taskId)
}

// Copy 复制任务
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @return 任务实体
// @return error 错误信息
func (taskDomain *TaskDomainImpl) Copy(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	// 1. 获取原任务（含所有权校验）
	existingTask, err := taskDomain.taskRepo.GetById(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}
	// 2. 将 *time.Time 转换为 sql.NullTime
	var startAt sql.NullTime
	if existingTask.StartAt != nil {
		startAt = sql.NullTime{Time: *existingTask.StartAt, Valid: true}
	}
	var endAt sql.NullTime
	if existingTask.EndAt != nil {
		endAt = sql.NullTime{Time: *existingTask.EndAt, Valid: true}
	}
	// 3. 转换提醒时间
	var remindAt sql.NullTime
	if existingTask.RemindAt != nil {
		remindAt = sql.NullTime{Time: *existingTask.RemindAt, Valid: true}
	}
	// 4. 构建创建任务值对象
	createTaskVO, err := valueobjects.NewCreateTask(
		0,
		existingTask.Name+"的复制",
		existingTask.Description,
		existingTask.State,
		existingTask.Priority,
		startAt,
		endAt,
		existingTask.ProjectId,
		existingTask.Tags,
		remindAt,
		existingTask.RemindRepeat,
		existingTask.RemindTime,
		existingTask.RemindWeekdays,
	)
	if err != nil {
		return nil, err
	}
	// 5. 创建新任务
	return taskDomain.taskRepo.Create(ctx, userId, createTaskVO)
}

// List 获取任务列表
// @param ctx 上下文
// @param userId 用户ID
// @param query 查询参数
// @param pagination 分页参数
// @return 任务实体列表
// @return error 错误
func (taskDomain *TaskDomainImpl) List(
	ctx context.Context,
	userId int64,
	query *valueobjects.QueryTask,
	pagination *valueobjects.Pagination,
) ([]*entities.Task, *valueobjects.Pagination, error) {
	// 1. 补充 query
	query.UserId = userId
	// 2. 构建查询句柄
	tx, err := taskDomain.taskRepo.BuildQueryTx(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	pagination.Page = query.Page
	pagination.Limit = query.Limit
	// 3. 查询任务列表
	taskEntities, pagination, err := taskDomain.taskRepo.ListWithQueryTx(ctx, tx, pagination)
	if err != nil {
		return nil, nil, err
	}
	return taskEntities, pagination, nil
}

// Snooze 稍后提醒
// @param ctx 上下文
// @param userId 用户ID
// @param taskId 任务ID
// @param durationMinutes 延迟分钟数
// @return 新提醒时间
// @return error 错误
func (taskDomain *TaskDomainImpl) Snooze(
	ctx context.Context,
	userId int64,
	taskId int64,
	durationMinutes int,
) (string, error) {
	// 1. 校验任务归属
	_, err := taskDomain.taskRepo.GetById(ctx, userId, taskId)
	if err != nil {
		return "", err
	}
	// 2. 计算新提醒时间
	newRemindAt := time.Now().Add(time.Duration(durationMinutes) * time.Minute).Format(time.RFC3339)
	// 3. 更新提醒时间
	err = taskDomain.taskRepo.Snooze(ctx, userId, taskId, newRemindAt)
	if err != nil {
		return "", err
	}
	return newRemindAt, nil
}

// CalculateNextRemindAt 计算下一次提醒时间
// @param remindAt 当前提醒时间
// @param repeat 重复类型
// @param remindTime 提醒时刻 HH:mm
// @param remindWeekdays 每周提醒位掩码
// @param endAt 任务结束时间
// @return *time.Time 下一次提醒时间，nil 表示无需重复
func CalculateNextRemindAt(
	remindAt time.Time,
	repeat int8,
	remindTime string,
	remindWeekdays int8,
	endAt *time.Time,
) *time.Time {
	// 1. 解析提醒时刻
	var hour, minute int
	if remindTime != "" {
		parts := strings.Split(remindTime, ":")
		if len(parts) == 2 {
			h, _ := strconv.Atoi(parts[0])
			m, _ := strconv.Atoi(parts[1])
			if h >= 0 && h <= 23 && m >= 0 && m <= 59 {
				hour = h
				minute = m
			}
		}
	}
	// 2. 计算目标时刻
	target := time.Date(
		remindAt.Year(),
		remindAt.Month(),
		remindAt.Day(),
		hour,
		minute,
		0,
		0,
		remindAt.Location(),
	)
	var next time.Time
	switch repeat {
	case 1: // daily
		next = target.AddDate(0, 0, 1)
	case 2: // weekly
		next = calculateNextWeekly(target, remindWeekdays)
		if next.IsZero() {
			return nil
		}
	case 3: // monthly
		next = target.AddDate(0, 1, 0)
	default:
		return nil
	}
	// 3. 如果有结束时间且超出则返回 nil
	if endAt != nil && next.After(*endAt) {
		return nil
	}
	return &next
}

// calculateNextWeekly 计算下一个匹配的星期
// @param from 起始时间
// @param weekdays 位掩码
// @return time.Time
func calculateNextWeekly(from time.Time, weekdays int8) time.Time {
	for i := 1; i <= 7; i++ {
		candidate := from.AddDate(0, 0, i)
		bit := int8(1 << uint(candidate.Weekday()))
		if weekdays&bit != 0 {
			return candidate
		}
	}
	return time.Time{}
}
