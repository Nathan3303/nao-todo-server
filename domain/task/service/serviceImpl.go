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

func NewTaskDomain(taskRepo repositories.Task) TaskDomain {
	return &TaskDomainImpl{taskRepo: taskRepo}
}

// === Task ===

func (d *TaskDomainImpl) GetById(ctx context.Context, userId int64, taskId int64) (*entities.Task, error) {
	return d.taskRepo.GetById(ctx, userId, taskId)
}

func (d *TaskDomainImpl) Create(ctx context.Context, userId int64, vo *valueobjects.CreateTask) (*entities.Task, error) {
	return d.taskRepo.Create(ctx, userId, vo)
}

func (d *TaskDomainImpl) Update(ctx context.Context, userId int64, taskId int64, vo *valueobjects.UpdateTask) error {
	return d.taskRepo.Update(ctx, userId, taskId, vo)
}

func (d *TaskDomainImpl) Delete(ctx context.Context, userId int64, taskId int64) error {
	return d.taskRepo.Delete(ctx, userId, taskId)
}

func (d *TaskDomainImpl) Restore(ctx context.Context, userId int64, taskId int64) error {
	return d.taskRepo.Restore(ctx, userId, taskId)
}

func (d *TaskDomainImpl) Copy(ctx context.Context, userId int64, taskId int64) (*entities.Task, error) {
	existingTask, err := d.taskRepo.GetById(ctx, userId, taskId)
	if err != nil {
		return nil, err
	}
	createTaskVO, err := valueobjects.NewCreateTask(
		0,
		existingTask.Name+"的复制",
		existingTask.Description,
		existingTask.State,
		existingTask.Priority,
		existingTask.StartAt,
		existingTask.EndAt,
		existingTask.ProjectId,
		existingTask.Tags,
		existingTask.RemindAt,
		existingTask.RemindRepeat,
		existingTask.RemindTime,
		existingTask.RemindWeekdays,
	)
	if err != nil {
		return nil, err
	}
	return d.taskRepo.Create(ctx, userId, createTaskVO)
}

func (d *TaskDomainImpl) List(ctx context.Context, userId int64, query *valueobjects.QueryTask, pagination *valueobjects.Pagination) ([]*entities.Task, *valueobjects.Pagination, error) {
	query.UserId = userId
	query.Page = pagination.Page
	query.Limit = pagination.Limit
	return d.taskRepo.List(ctx, userId, query, pagination)
}

func (d *TaskDomainImpl) Snooze(ctx context.Context, userId int64, taskId int64, durationMinutes int) (string, error) {
	if _, err := d.taskRepo.GetById(ctx, userId, taskId); err != nil {
		return "", err
	}
	newRemindAt := time.Now().Add(time.Duration(durationMinutes) * time.Minute).Format(time.RFC3339)
	if err := d.taskRepo.Snooze(ctx, userId, taskId, newRemindAt); err != nil {
		return "", err
	}
	return newRemindAt, nil
}

func (d *TaskDomainImpl) ProcessReminders(ctx context.Context) ([]*entities.Task, error) {
	tasks, err := d.taskRepo.GetDueReminders(ctx)
	if err != nil {
		return nil, err
	}
	for _, task := range tasks {
		if task.RemindRepeat != 0 {
			next := d.calculateNextRemindAt(*task.RemindAt, task.RemindRepeat, task.RemindTime, task.RemindWeekdays, task.EndAt)
			if next != nil {
				d.taskRepo.UpdateRemindAt(ctx, task.Id, next.Format("2006-01-02 15:04:05"))
			} else {
				d.taskRepo.ClearRemindRepeat(ctx, task.Id)
			}
		} else {
			d.taskRepo.UpdateRemindAt(ctx, task.Id, "")
		}
	}
	return tasks, nil
}

func (d *TaskDomainImpl) calculateNextRemindAt(remindAt time.Time, repeat int8, remindTime string, remindWeekdays int8, endAt *time.Time) *time.Time {
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
	target := time.Date(remindAt.Year(), remindAt.Month(), remindAt.Day(), hour, minute, 0, 0, remindAt.Location())
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

// === CheckItem ===

func (d *TaskDomainImpl) GetCheckItemById(ctx context.Context, userId, checkItemId int64) (*entities.CheckItem, error) {
	return d.taskRepo.GetCheckItemById(ctx, userId, checkItemId)
}

func (d *TaskDomainImpl) CreateCheckItem(ctx context.Context, userId int64, vo *valueobjects.CreateCheckItem) (*entities.CheckItem, error) {
	vo.SortId = d.taskRepo.GetMaxCheckItemSortId(ctx, userId, vo.TaskId) + 1
	return d.taskRepo.CreateCheckItem(ctx, userId, vo)
}

func (d *TaskDomainImpl) UpdateCheckItem(ctx context.Context, userId, checkItemId int64, vo *valueobjects.UpdateCheckItem) error {
	return d.taskRepo.UpdateCheckItem(ctx, userId, checkItemId, vo)
}

func (d *TaskDomainImpl) DeleteCheckItem(ctx context.Context, userId, checkItemId int64) error {
	return d.taskRepo.DeleteCheckItem(ctx, userId, checkItemId)
}

func (d *TaskDomainImpl) ListCheckItems(ctx context.Context, userId, taskId int64) ([]*entities.CheckItem, error) {
	return d.taskRepo.ListCheckItems(ctx, userId, taskId)
}

func (d *TaskDomainImpl) BatchUpdateCheckItems(ctx context.Context, userId int64, vos []*valueobjects.BatchUpdateCheckItem) ([]*entities.CheckItem, error) {
	return d.taskRepo.BatchUpdateCheckItems(ctx, userId, vos)
}

// === Comment ===

func (d *TaskDomainImpl) GetCommentById(ctx context.Context, userId, commentId int64) (*entities.Comment, error) {
	return d.taskRepo.GetCommentById(ctx, userId, commentId)
}

func (d *TaskDomainImpl) CreateComment(ctx context.Context, userId int64, vo *valueobjects.CreateComment) (*entities.Comment, error) {
	return d.taskRepo.CreateComment(ctx, userId, vo)
}

func (d *TaskDomainImpl) UpdateComment(ctx context.Context, userId, commentId int64, vo *valueobjects.UpdateComment) error {
	return d.taskRepo.UpdateComment(ctx, userId, commentId, vo)
}

func (d *TaskDomainImpl) DeleteComment(ctx context.Context, userId, commentId int64) error {
	return d.taskRepo.DeleteComment(ctx, userId, commentId)
}

func (d *TaskDomainImpl) ListComments(ctx context.Context, userId, taskId int64) ([]*entities.Comment, error) {
	return d.taskRepo.ListComments(ctx, userId, taskId)
}

func (d *TaskDomainImpl) SyncCommentUserProfile(ctx context.Context, userId int64, nickname, avatar string) error {
	return d.taskRepo.SyncCommentUserProfile(ctx, userId, nickname, avatar)
}
