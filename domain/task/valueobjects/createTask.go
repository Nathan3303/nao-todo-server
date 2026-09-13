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
// 客户端已**提供** startAt（含显式空串清空）时不得覆盖（DEF-SYNC-04 / SYNC-DEF-01）：
// 显式空串（Valid=true, IsNull=true）是「清空」意图，不能被 fallback 复活为 now；
// 仅当 startAt 字段**缺省**（Valid=false）时按既有两段式规则兜底：
//   - endAt 有效 ⇒ 继承规则（startAt = 服务端 now；endAt 已过去则 now−1min）
//   - endAt 缺省/清空 ⇒ startAt/endAt 保持原样（未安排）
func (createTask *CreateTask) FillStartAt() {
	// 已提供（Valid=true，含显式清空）⇒ 原样保留，不得用服务端 now 覆盖
	if createTask.StartAt.Valid {
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

// nullableTimeFromCreateReq 创建/推送请求中的可空时间字段归一为 NullableTime：
//   - nil（字段缺省，反序列化后与 JSON null 不可区分）⇒ 未提供（Valid=false）
//   - ""（显式空串）⇒ 显式清空（Valid=true, IsNull=true）
//   - 合法时间串 ⇒ 指定值
//   - 非法非空串 ⇒ 视同未提供（沿用历史「无效即缺失」语义）
func nullableTimeFromCreateReq(s *string) types.NullableTime {
	if s == nil {
		return types.NewNullableTimeNull()
	}
	if *s == "" {
		return types.NewNullableTimeSetToNull()
	}
	nt := types.NewNullableTimeByTimeStr(*s)
	if nt.IsNull {
		return types.NewNullableTimeNull()
	}
	return nt
}

// NewCreateTask 创建创建任务值对象
// @param parentTaskId 父任务 ID
// @param name 任务名称
// @param description 任务描述
// @param state 任务状态
// @param priority 任务优先级
// @param startAt 任务开始时间（nil=缺省，""=清空）
// @param endAt 任务结束时间（nil=缺省，""=清空）
// @param projectId 项目 ID
// @param tags 任务标签
// @param remindAt 提醒时间（nil=缺省，""=清空）
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
	startAt *string,
	endAt *string,
	projectId types.ProjectID,
	tags []string,
	remindAt *string,
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
		StartAt:        nullableTimeFromCreateReq(startAt),
		EndAt:          nullableTimeFromCreateReq(endAt),
		ProjectId:      projectId,
		Tags:           tags,
		RemindAt:       nullableTimeFromCreateReq(remindAt),
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
