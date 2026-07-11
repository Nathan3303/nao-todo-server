package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/repositories"
	"naotodoserver/domain/task/valueobjects"
)

// NewTaskDomain 任务域实现
func NewTaskDomain(taskRepo repositories.Task) TaskDomain {
	return &TaskDomainImpl{taskRepo: taskRepo}
}

// CreateTask 创建任务
func (d *TaskDomainImpl) CreateTask(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTask,
) (*entities.Task, error) {
	vo.SortId = d.taskRepo.GetMaxSortId(ctx, userId) + 1
	return d.taskRepo.Create(ctx, userId, vo)
}

// Copy 复制任务
func (d *TaskDomainImpl) Copy(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	// 检查任务是否存在
	task, err := d.taskRepo.GetById(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}
	// 创建新任务VO
	var vo valueobjects.CreateTask
	vo.ParentTaskId = task.ParentTaskId
	vo.Name = task.Name + "的复制"
	vo.Description = task.Description
	vo.State = task.State
	vo.Priority = task.Priority
	vo.StartAt = task.StartAt
	vo.EndAt = task.EndAt
	vo.ProjectId = task.ProjectId
	vo.Tags = task.Tags
	// vo.RemindAt = task.RemindAt
	// vo.RemindRepeat = task.RemindRepeat
	// vo.RemindTime = task.RemindTime
	// vo.RemindWeekdays = task.RemindWeekdays
	if vo.Validate() != nil {
		return nil, err
	}
	// 创建新任务并返回新任务实体
	return d.CreateTask(ctx, userId, &vo)
}

// List 获取任务列表
func (d *TaskDomainImpl) List(
	ctx context.Context,
	userId int64,
	query *valueobjects.QueryTask,
	pagination *valueobjects.Pagination,
) ([]*entities.Task, *valueobjects.Pagination, error) {
	query.UserId = userId
	query.Page = pagination.Page
	query.Limit = pagination.Limit
	return d.taskRepo.List(ctx, userId, query, pagination)
}

// --- 检查项相关 ---

// CreateCheckItem 创建检查项
func (d *TaskDomainImpl) CreateCheckItem(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskCheckItem,
) (*entities.TaskCheckItem, error) {
	vo.SortId = d.taskRepo.GetMaxCheckItemSortId(ctx, userId, vo.TaskId) + 1
	return d.taskRepo.CreateCheckItem(ctx, userId, vo)
}

// --- 任务提醒相关 ---

// Snooze 延迟任务
func (d *TaskDomainImpl) Snooze(
	ctx context.Context,
	userId int64,
	taskId int64,
	durationMinutes int,
) (string, error) {
	if _, err := d.taskRepo.GetById(ctx, userId, taskId); err != nil {
		return "", err
	}
	newRemindAt := time.
		Now().
		Add(time.Duration(durationMinutes) * time.Minute).
		Format(time.RFC3339)
	if err := d.taskRepo.Snooze(ctx, userId, taskId, newRemindAt); err != nil {
		return "", err
	}
	return newRemindAt, nil
}

// ProcessReminders 处理任务提醒
func (d *TaskDomainImpl) ProcessReminders(ctx context.Context) ([]*entities.Task, error) {
	// 获取所有待提醒任务
	tasks, err := d.taskRepo.GetDueReminders(ctx)
	if err != nil {
		return nil, err
	}
	// 处理每个任务的提醒时间
	for _, task := range tasks {
		if task.RemindRepeat != 0 {
			next := d.calculateNextRemindAt(
				task.RemindAt.Time,
				task.RemindRepeat,
				task.RemindTime,
				task.RemindWeekdays,
				&task.EndAt.Time,
			)
			if next != nil {
				err := d.taskRepo.UpdateRemindAt(
					ctx,
					task.Id,
					next.Format(time.RFC3339),
				)
				if err != nil {
					return nil, err
				}
			} else {
				err := d.taskRepo.ClearRemindRepeat(ctx, task.Id)
				if err != nil {
					return nil, err
				}
			}
		} else {
			err := d.taskRepo.UpdateRemindAt(ctx, task.Id, "")
			if err != nil {
				return nil, err
			}
		}
	}
	return tasks, nil
}

// calculateNextRemindAt 计算下一个提醒时间
func (d *TaskDomainImpl) calculateNextRemindAt(
	remindAt time.Time,
	repeat uint8,
	remindTime string,
	remindWeekdays uint8,
	endAt *time.Time,
) *time.Time {
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
	case 1:
		next = target.AddDate(0, 0, 1)
	case 2:
		next = calculateNextWeekly(target, remindWeekdays)
		if next.IsZero() {
			return nil
		}
	case 3:
		next = target.AddDate(0, 1, 0)
	default:
		return nil
	}
	if endAt != nil && next.After(*endAt) {
		return nil
	}
	return &next
}

// calculateNextWeekly 计算下一个周日
func calculateNextWeekly(from time.Time, weekdays uint8) time.Time {
	for i := 1; i <= 7; i++ {
		candidate := from.AddDate(0, 0, i)
		bit := uint8(1 << candidate.Weekday())
		if weekdays&bit != 0 {
			return candidate
		}
	}
	return time.Time{}
}
