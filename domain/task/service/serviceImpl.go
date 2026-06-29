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

// GetById 获取任务详情
func (d *TaskDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	return d.taskRepo.GetById(ctx, userId, taskId)
}

// Create 创建任务
func (d *TaskDomainImpl) Create(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTask,
) (*entities.Task, error) {
	return d.taskRepo.Create(ctx, userId, vo)
}

// Update 更新任务
func (d *TaskDomainImpl) Update(
	ctx context.Context,
	userId int64,
	taskId int64,
	vo *valueobjects.UpdateTask,
) error {
	return d.taskRepo.Update(ctx, userId, taskId, vo)
}

// Delete 删除任务
func (d *TaskDomainImpl) Delete(ctx context.Context, userId int64, taskId int64) error {
	return d.taskRepo.Delete(ctx, userId, taskId)
}

// Restore 恢复任务
func (d *TaskDomainImpl) Restore(ctx context.Context, userId int64, taskId int64) error {
	return d.taskRepo.Restore(ctx, userId, taskId)
}

// Copy 复制任务
func (d *TaskDomainImpl) Copy(
	ctx context.Context,
	userId int64,
	taskId int64,
) (*entities.Task, error) {
	// 检查任务是否存在
	existingTask, err := d.taskRepo.GetById(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}
	// 创建新任务VO
	createTaskVO, err := valueobjects.NewCreateTask(
		0,
		existingTask.Name+"的复制",
		existingTask.Description,
		existingTask.State,
		existingTask.Priority,
		&existingTask.StartAt.Time,
		&existingTask.EndAt.Time,
		existingTask.ProjectId,
		existingTask.Tags,
		&existingTask.RemindAt.Time,
		existingTask.RemindRepeat,
		existingTask.RemindTime,
		existingTask.RemindWeekdays,
	)
	if err != nil {
		return nil, err
	}
	return d.taskRepo.Create(ctx, userId, createTaskVO)
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

// GetCheckItemById 获取检查项详情
func (d *TaskDomainImpl) GetCheckItemById(
	ctx context.Context,
	userId int64,
	checkItemId int64,
) (*entities.TaskCheckItem, error) {
	return d.taskRepo.GetCheckItemById(ctx, userId, checkItemId)
}

// CreateCheckItem 创建检查项
func (d *TaskDomainImpl) CreateCheckItem(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskCheckItem,
) (*entities.TaskCheckItem, error) {
	vo.SortId = d.taskRepo.GetMaxCheckItemSortId(ctx, userId, vo.TaskId) + 1
	return d.taskRepo.CreateCheckItem(ctx, userId, vo)
}

// UpdateCheckItem 更新检查项
func (d *TaskDomainImpl) UpdateCheckItem(
	ctx context.Context,
	userId int64,
	checkItemId int64,
	vo *valueobjects.UpdateTaskCheckItem,
) error {
	return d.taskRepo.UpdateCheckItem(ctx, userId, checkItemId, vo)
}

// DeleteCheckItem 删除检查项
func (d *TaskDomainImpl) DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error {
	return d.taskRepo.DeleteCheckItem(ctx, userId, checkItemId)
}

// ListCheckItems 获取检查项列表
func (d *TaskDomainImpl) ListCheckItems(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.TaskCheckItem, error) {
	return d.taskRepo.ListCheckItems(ctx, userId, taskId)
}

// BatchUpdateCheckItems 批量更新检查项
func (d *TaskDomainImpl) BatchUpdateCheckItems(
	ctx context.Context,
	userId int64,
	vos []*valueobjects.BatchUpdateTaskCheckItem,
) ([]*entities.TaskCheckItem, error) {
	return d.taskRepo.BatchUpdateCheckItems(ctx, userId, vos)
}

// --- 评论相关 ---

// GetCommentById 获取评论详情
func (d *TaskDomainImpl) GetCommentById(
	ctx context.Context,
	userId int64,
	commentId int64,
) (*entities.TaskComment, error) {
	return d.taskRepo.GetCommentById(ctx, userId, commentId)
}

// CreateComment 创建评论
func (d *TaskDomainImpl) CreateComment(
	ctx context.Context,
	userId int64,
	vo *valueobjects.CreateTaskComment,
) (*entities.TaskComment, error) {
	return d.taskRepo.CreateComment(ctx, userId, vo)
}

// UpdateComment 更新评论
func (d *TaskDomainImpl) UpdateComment(
	ctx context.Context,
	userId int64,
	commentId int64,
	vo *valueobjects.UpdateTaskComment,
) error {
	return d.taskRepo.UpdateComment(ctx, userId, commentId, vo)
}

// DeleteComment 删除评论
func (d *TaskDomainImpl) DeleteComment(ctx context.Context, userId, commentId int64) error {
	return d.taskRepo.DeleteComment(ctx, userId, commentId)
}

// ListComments 获取评论列表
func (d *TaskDomainImpl) ListComments(
	ctx context.Context,
	userId int64,
	taskId int64,
) ([]*entities.TaskComment, error) {
	return d.taskRepo.ListComments(ctx, userId, taskId)
}

// SyncCommentUserProfile 同步评论用户配置
func (d *TaskDomainImpl) SyncCommentUserProfile(
	ctx context.Context,
	userId int64,
	nickname string,
	avatar string,
) error {
	return d.taskRepo.SyncCommentUserProfile(ctx, userId, nickname, avatar)
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
				d.taskRepo.UpdateRemindAt(
					ctx,
					task.Id,
					next.Format(time.RFC3339),
				)
			} else {
				d.taskRepo.ClearRemindRepeat(ctx, task.Id)
			}
		} else {
			d.taskRepo.UpdateRemindAt(ctx, task.Id, "")
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
		bit := uint8(1 << uint(candidate.Weekday()))
		if weekdays&bit != 0 {
			return candidate
		}
	}
	return time.Time{}
}
