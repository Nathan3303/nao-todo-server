package valueobjects

import (
	"errors"
	"strings"

	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// UpdateTask 更新任务值对象
type UpdateTask struct {
	UserId         int64
	ParentTaskId   *int64
	Name           *string
	Description    *string
	State          *uint8
	Priority       *uint8
	StartAt        types.NullableTime
	EndAt          types.NullableTime
	ProjectId      *int64
	Tags           []string
	ArchivedAt     types.NullableTime
	StarMarkAt     types.NullableTime
	GivenUpAt      types.NullableTime
	RemindAt       types.NullableTime
	RemindRepeat   *uint8
	RemindTime     *string
	RemindWeekdays *uint8
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
func NewUpdateTask(
	userId int64,
	parentTaskId *int64,
	name *string,
	description *string,
	state *uint8,
	priority *uint8,
	startAt *string,
	endAt *string,
	projectId *int64,
	tags []string,
	archivedAt *string,
	starMarkAt *string,
	givenUpAt *string,
	remindAt *string,
	remindRepeat *uint8,
	remindTime *string,
	remindWeekdays *uint8,
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
	vo.RemindAt = types.NewNullableTimeByTimeStrPtr(remindAt)
	vo.RemindRepeat = remindRepeat
	vo.RemindTime = remindTime
	vo.RemindWeekdays = remindWeekdays
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
