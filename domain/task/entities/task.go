package entities

import (
	"errors"
	"time"
)

type Task struct {
	Id           int64
	UserId       int64
	ProjectId    int64
	ParentTodoId int64
	Name         string
	Description  string
	State        int8
	Priority     int8
	StartAt      *time.Time
	EndAt        *time.Time
	ArchivedAt   *time.Time
	StarMarkAt   *time.Time
	GivenUpAt    *time.Time
	Tags         []string
	UpdatedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    time.Time
}

func (task *Task) GetFormatedStartAt() string {
	if task.StartAt == nil {
		return ""
	}
	return task.StartAt.Format(time.RFC3339)
}

func (task *Task) GetFormatedEndAt() string {
	if task.EndAt == nil {
		return ""
	}
	return task.EndAt.Format(time.RFC3339)
}

func (task *Task) ParseArchivedAt() (string, bool) {
	if task.ArchivedAt == nil {
		return "", false
	}
	return task.ArchivedAt.Format(time.RFC3339), true
}

func (task *Task) ParseStarMarkAt() (string, bool) {
	if task.StarMarkAt == nil {
		return "", false
	}
	return task.StarMarkAt.Format(time.RFC3339), true
}

func (task *Task) ParseGivenUpAt() (string, bool) {
	if task.GivenUpAt == nil {
		return "", false
	}
	return task.GivenUpAt.Format(time.RFC3339), true
}

func (task *Task) GetFormatedUpdatedAt() string {
	if task.UpdatedAt.IsZero() {
		return ""
	}
	return task.UpdatedAt.Format(time.RFC3339)
}

func (task *Task) GetFormatedCreatedAt() string {
	if task.CreatedAt.IsZero() {
		return ""
	}
	return task.CreatedAt.Format(time.RFC3339)
}

func (task *Task) GetFormatedDeletedAt() (string, bool) {
	if task.DeletedAt.IsZero() {
		return "", false
	}
	return task.DeletedAt.Format(time.RFC3339), true
}

func (task *Task) IsEndAtValid() bool {
	if task.EndAt == nil {
		return false
	}
	if task.StartAt == nil {
		return true
	}
	return task.EndAt.After(*task.StartAt)
}

func (task *Task) IsArchivedAtValid() bool {
	if task.ArchivedAt == nil {
		return true
	}
	if task.StartAt == nil {
		return true
	}
	return task.ArchivedAt.After(*task.StartAt)
}

func (task *Task) IsStarMarkAtValid() bool {
	if task.StarMarkAt == nil {
		return true
	}
	if task.StartAt == nil {
		return true
	}
	return task.StarMarkAt.After(*task.StartAt)
}

func (task *Task) IsGivenUpAtValid() bool {
	if task.GivenUpAt == nil {
		return true
	}
	if task.StartAt == nil {
		return true
	}
	return task.GivenUpAt.After(*task.StartAt)
}

func (task *Task) IsDatesValid() error {
	if !task.IsEndAtValid() {
		return errors.New("时间参数无效 - 结束时间必须晚于开始时间")
	}
	if !task.IsArchivedAtValid() {
		return errors.New("时间参数无效 - 归档时间必须晚于开始时间")
	}
	if !task.IsStarMarkAtValid() {
		return errors.New("时间参数无效 - 收藏时间必须晚于开始时间")
	}
	if !task.IsGivenUpAtValid() {
		return errors.New("时间参数无效 - 放弃时间必须晚于开始时间")
	}
	return nil
}
