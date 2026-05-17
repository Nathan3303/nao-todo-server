package valueobjects

import (
	"database/sql"
	"errors"
	"time"

	"naotodoserver/infrastructure/utils"
)

// CreateTask 创建任务值对象
type CreateTask struct {
	ParentTaskId int64
	Name         string
	Description  string
	State        int8
	Priority     int8
	StartAt      sql.NullTime
	EndAt        sql.NullTime
	ProjectId    int64
	Tags         []string
}

// Validate 验证创建任务值对象
// @return error 验证失败返回错误
func (createTask *CreateTask) Validate() error {
	if createTask.Name == "" {
		return errors.New("任务名称不能为空")
	}
	if utils.RuneLength(createTask.Name) > 256 {
		return errors.New("任务名称最多256个字符")
	}
	if createTask.Description != "" && utils.RuneLength(createTask.Description) > 512 {
		return errors.New("任务描述最多512个字符")
	}
	return nil
}

// FillStartAt 填充开始时间
func (createTask *CreateTask) FillStartAt() {
	if createTask.StartAt.Valid || !createTask.EndAt.Valid {
		return
	}
	t := time.Now()
	if createTask.EndAt.Time.Before(t) {
		createTask.StartAt = sql.NullTime{Time: t.Add(-1 * time.Minute), Valid: true}
	} else {
		createTask.StartAt = sql.NullTime{Time: t, Valid: true}
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
// @return *CreateTask 创建任务值对象
// @return error 创建失败返回错误
func NewCreateTask(
	parentTaskId int64,
	name string,
	description string,
	state int8,
	priority int8,
	startAt sql.NullTime,
	endAt sql.NullTime,
	projectId int64,
	tags []string,
) (*CreateTask, error) {
	createTask := &CreateTask{
		ParentTaskId: parentTaskId,
		Name:         name,
		Description:  description,
		State:        state,
		Priority:     priority,
		StartAt:      startAt,
		EndAt:        endAt,
		ProjectId:    projectId,
		Tags:         tags,
	}
	createTask.FillStartAt()
	err := createTask.Validate()
	if err != nil {
		return nil, err
	}
	return createTask, nil
}
