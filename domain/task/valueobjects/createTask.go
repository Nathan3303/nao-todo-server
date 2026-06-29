package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/textutils"
)

// CreateTask 创建任务值对象
type CreateTask struct {
	ParentTaskId   int64
	Name           string
	Description    string
	State          uint8
	Priority       uint8
	StartAt        *time.Time
	EndAt          *time.Time
	ProjectId      int64
	Tags           []string
	RemindAt       *time.Time
	RemindRepeat   uint8
	RemindTime     string
	RemindWeekdays uint8
}

// Validate 验证创建任务值对象
// @return error 验证失败返回错误
func (createTask *CreateTask) Validate() error {
	if createTask.Name == "" {
		return errors.New("任务名称不能为空")
	}
	if textutils.RuneLength(createTask.Name) > 256 {
		return errors.New("任务名称最多256个字符")
	}
	if createTask.Description != "" && textutils.RuneLength(createTask.Description) > 512 {
		return errors.New("任务描述最多512个字符")
	}
	return nil
}

// FillStartAt 填充开始时间
func (createTask *CreateTask) FillStartAt() {
	if createTask.StartAt != nil || createTask.EndAt == nil {
		return
	}
	t := time.Now()
	if createTask.EndAt.Before(t) {
		before := t.Add(-1 * time.Minute)
		createTask.StartAt = &before
	} else {
		createTask.StartAt = &t
	}
}

// NewCreateTask 创建创建任务值对象
// @param parentTaskId 父任务 ID
// @param name 任务名称
// @param description 任务描述
// @param state 任务状态
// @param priority 任务优先级
// @param startAt 任务开始时间
// @param endAt 任务结束时间
// @param projectId 项目 ID
// @param tags 任务标签
// @param remindAt 提醒时间
// @param remindRepeat 重复提醒类型
// @param remindTime 提醒时刻
// @param remindWeekdays 每周提醒星期
// @return *CreateTask 创建任务值对象
// @return error 创建失败返回错误
func NewCreateTask(
	parentTaskId int64,
	name string,
	description string,
	state uint8,
	priority uint8,
	startAt *time.Time,
	endAt *time.Time,
	projectId int64,
	tags []string,
	remindAt *time.Time,
	remindRepeat uint8,
	remindTime string,
	remindWeekdays uint8,
) (*CreateTask, error) {
	createTask := &CreateTask{
		ParentTaskId:   parentTaskId,
		Name:           name,
		Description:    description,
		State:          state,
		Priority:       priority,
		StartAt:        startAt,
		EndAt:          endAt,
		ProjectId:      projectId,
		Tags:           tags,
		RemindAt:       remindAt,
		RemindRepeat:   remindRepeat,
		RemindTime:     remindTime,
		RemindWeekdays: remindWeekdays,
	}
	createTask.FillStartAt()
	err := createTask.Validate()
	if err != nil {
		return nil, err
	}
	return createTask, nil
}
