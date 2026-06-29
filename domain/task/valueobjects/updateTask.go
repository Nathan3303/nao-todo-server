package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

type UpdateTask struct {
	UserId         int64
	ParentTaskId   *int64
	Name           *string
	Description    *string
	State          *uint8
	Priority       *uint8
	StartAt        *types.NullableTime
	EndAt          *types.NullableTime
	ProjectId      *int64
	Tags           []string
	ArchivedAt     *types.NullableTime
	StarMarkAt     *types.NullableTime
	GivenUpAt      *types.NullableTime
	RemindAt       *types.NullableTime
	RemindRepeat   *uint8
	RemindTime     *string
	RemindWeekdays *uint8
}

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
	if updateTask.EndAt != nil &&
		updateTask.EndAt.ShouldUpdate() &&
		!updateTask.EndAt.IsSetToNull() &&
		updateTask.StartAt != nil &&
		updateTask.StartAt.ShouldUpdate() &&
		!updateTask.StartAt.IsSetToNull() {
		endTime, endOk := updateTask.EndAt.Value()
		startTime, startOk := updateTask.StartAt.Value()
		if endOk && startOk && endTime.Before(startTime) {
			return errors.New("结束时间不能早于开始时间")
		}
	}
	return nil
}

func NewUpdateTask(
	userId int64,
	parentTaskId *int64,
	name *string,
	description *string,
	state *uint8,
	priority *uint8,
	startAt *types.NullableTime,
	endAt *types.NullableTime,
	projectId *int64,
	tags []string,
	archivedAt *types.NullableTime,
	starMarkAt *types.NullableTime,
	givenUpAt *types.NullableTime,
	remindAt *types.NullableTime,
	remindRepeat *uint8,
	remindTime *string,
	remindWeekdays *uint8,
) (*UpdateTask, error) {
	updateTask := &UpdateTask{
		UserId:         userId,
		ParentTaskId:   parentTaskId,
		Name:           name,
		Description:    description,
		State:          state,
		Priority:       priority,
		StartAt:        startAt,
		EndAt:          endAt,
		ProjectId:      projectId,
		Tags:           tags,
		ArchivedAt:     archivedAt,
		StarMarkAt:     starMarkAt,
		GivenUpAt:      givenUpAt,
		RemindAt:       remindAt,
		RemindRepeat:   remindRepeat,
		RemindTime:     remindTime,
		RemindWeekdays: remindWeekdays,
	}
	err := updateTask.Validate()
	if err != nil {
		return nil, err
	}
	return updateTask, nil
}
