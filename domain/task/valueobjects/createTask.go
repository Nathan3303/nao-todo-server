package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// CreateTask 创建任务值对象
type CreateTask struct {
	Id             int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      types.NullableTime
	ParentTaskId   types.TaskID
	Name           string
	Description    string
	State          entities.TaskState
	Priority       entities.TaskPriority
	StartAt        types.NullableTime
	EndAt          types.NullableTime
	ProjectId      types.ProjectID
	Tags           []string
	RemindAt       types.NullableTime
	RemindRepeat   uint8
	RemindTime     string
	RemindWeekdays uint8
	SortId         uint16
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
// 客户端已显式提供有效 startAt（非空/非 null）时**不得覆盖**（DEF-SYNC-04：创建/推送往返秒级瞬时不变）；
// 仅当 startAt 缺失/无效时按既有两段式规则兜底：
//   - endAt 有效 ⇒ 继承规则（startAt = 服务端 now；endAt 已过去则 now−1min）
//   - endAt 缺失/无效 ⇒ startAt/endAt 保持原样（未安排）
func (createTask *CreateTask) FillStartAt() {
	// 客户端已提供有效 startAt ⇒ 原样保留，不得用服务端 now 覆盖
	if !createTask.StartAt.IsNull {
		return
	}
	if createTask.EndAt.IsNull {
		return
	}
	t := time.Now()
	if createTask.EndAt.Time.Before(t) {
		before := t.Add(-1 * time.Minute)
		createTask.StartAt.SetTime(before)
	} else {
		createTask.StartAt.SetTime(t)
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
	parentTaskId types.TaskID,
	name string,
	description string,
	state entities.TaskState,
	priority entities.TaskPriority,
	startAt string,
	endAt string,
	projectId types.ProjectID,
	tags []string,
	remindAt string,
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
		StartAt:        types.NewNullableTimeByTimeStr(startAt),
		EndAt:          types.NewNullableTimeByTimeStr(endAt),
		ProjectId:      projectId,
		Tags:           tags,
		RemindAt:       types.NewNullableTimeByTimeStr(remindAt),
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
