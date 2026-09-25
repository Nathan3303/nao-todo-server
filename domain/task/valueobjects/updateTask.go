package valueobjects

import (
	"errors"
	"strings"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// UpdateTask 更新任务值对象
type UpdateTask struct {
	UserId         int64
	ParentTaskId   *int64
	Name           *string
	Description    *string
	State          *entities.TaskState
	Priority       *entities.TaskPriority
	StartAt        types.NullableTime
	EndAt          types.NullableTime
	ProjectId      *int64
	Tags           []string
	ArchivedAt     types.NullableTime
	StarMarkAt     types.NullableTime
	GivenUpAt      types.NullableTime
	CompletedAt    types.NullableTime
	RemindAt       types.NullableTime
	RemindRepeat   *uint8
	RemindTime     *string
	RemindWeekdays *uint8
	SortId         *uint16
	// UpdatedAt 乐观锁时间戳（零值=未提供）：早于服务端当前版本时不更新（LWW）
	UpdatedAt time.Time
}

// Validate 验证更新任务值对象
func (updateTask *UpdateTask) Validate() error {
	if updateTask.Name != nil && textutils.RuneLength(*updateTask.Name) == 0 {
		return errors.New("任务名称不能为空")
	}
	if updateTask.Name != nil && textutils.RuneLength(*updateTask.Name) > 256 {
		return errors.New("任务名称最多256个字符")
	}
	if updateTask.Description != nil && textutils.RuneLength(*updateTask.Description) > 512 {
		return errors.New("任务描述最多512个字符")
	}
	if updateTask.EndAt.IsNull && updateTask.StartAt.IsNull {
		endTime, endOk := updateTask.EndAt.Value()
		startTime, startOk := updateTask.StartAt.Value()
		if endOk && startOk && endTime.Before(startTime) {
			return errors.New("结束时间不能早于开始时间")
		}
	}
	return nil
}

// Trim 去除任务名称和描述的首尾空格
func (updateTask *UpdateTask) Trim() error {
	if updateTask.Name != nil {
		*updateTask.Name = strings.TrimSpace(*updateTask.Name)
	}
	if updateTask.Description != nil {
		*updateTask.Description = strings.TrimSpace(*updateTask.Description)
	}
	return nil
}

// NewUpdateTask 创建更新任务值对象
// @param userId 用户 ID
// @param parentTaskId 父任务 ID
// @param name 任务名称
// @param description 任务描述
// @param state 任务状态
// @param priority 任务优先级
// @param startAt 开始时间
// @param endAt 结束时间
// @param projectId 项目 ID
// @param tags 标签
// @param archivedAt 归档时间
// @param starMarkAt 收藏时间
// @param givenUpAt 放弃时间
// @param completedAt 完成时间
// @param remindAt 提醒时间
// @param remindRepeat 提醒重复类型
// @param remindTime 提醒时间字符串
// @param remindWeekdays 提醒星期位掩码
// @param sortId 排序 ID
// @return *UpdateTask 更新任务值对象
// @return error 错误信息
func NewUpdateTask(
	userId int64,
	parentTaskId *int64,
	name *string,
	description *string,
	state *entities.TaskState,
	priority *entities.TaskPriority,
	startAt *string,
	endAt *string,
	projectId *int64,
	tags []string,
	archivedAt *string,
	starMarkAt *string,
	givenUpAt *string,
	completedAt *string,
	remindAt *string,
	remindRepeat *uint8,
	remindTime *string,
	remindWeekdays *uint8,
	sortId *uint16,
) (*UpdateTask, error) {
	var vo UpdateTask
	vo.UserId = userId
	vo.ParentTaskId = parentTaskId
	vo.Name = name
	vo.Description = description
	vo.State = state
	vo.Priority = priority
	vo.StartAt = types.NewNullableTimeByTimeStrPtr(startAt)
	vo.EndAt = types.NewNullableTimeByTimeStrPtr(endAt)
	vo.ProjectId = projectId
	vo.Tags = tags
	vo.ArchivedAt = types.NewNullableTimeByTimeStrPtr(archivedAt)
	vo.StarMarkAt = types.NewNullableTimeByTimeStrPtr(starMarkAt)
	vo.GivenUpAt = types.NewNullableTimeByTimeStrPtr(givenUpAt)
	vo.CompletedAt = types.NewNullableTimeByTimeStrPtr(completedAt)
	vo.RemindAt = types.NewNullableTimeByTimeStrPtr(remindAt)
	vo.RemindRepeat = remindRepeat
	vo.RemindTime = remindTime
	vo.RemindWeekdays = remindWeekdays
	vo.SortId = sortId
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	err = vo.Trim()
	if err != nil {
		return nil, err
	}
	return &vo, nil
}
