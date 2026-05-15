package valueobjects

import (
	"errors"
	"naotodoserver/infrastructure/utils"
)

type UpdateTask struct {
	UserId       int64
	ParentTaskId *int64
	Name         *string
	Description  *string
	State        *int8
	Priority     *int8
	StartAt      *utils.NullableTime
	EndAt        *utils.NullableTime
	ProjectId    *int64
	Tags         []string
	ArchivedAt   *utils.NullableTime
	StarMarkAt   *utils.NullableTime
	GivenUpAt    *utils.NullableTime
}

func (updateTask *UpdateTask) Validate() error {
	if updateTask.Name != nil && utils.RuneLength(*updateTask.Name) == 0 {
		return errors.New("任务名称不能为空")
	}
	if updateTask.Name != nil && utils.RuneLength(*updateTask.Name) > 256 {
		return errors.New("任务名称最多256个字符")
	}
	if updateTask.Description != nil && utils.RuneLength(*updateTask.Description) > 512 {
		return errors.New("任务描述最多512个字符")
	}
	if updateTask.EndAt != nil && updateTask.EndAt.ShouldUpdate() && !updateTask.EndAt.IsSetToNull() &&
		updateTask.StartAt != nil && updateTask.StartAt.ShouldUpdate() && !updateTask.StartAt.IsSetToNull() &&
		updateTask.EndAt.ToSqlNullTime().Time.Before(updateTask.StartAt.ToSqlNullTime().Time) {
		return errors.New("结束时间不能早于开始时间")
	}
	return nil
}

func NewUpdateTask(
	userId int64,
	parentTaskId *int64,
	name *string,
	description *string,
	state *int8,
	priority *int8,
	startAt *utils.NullableTime,
	endAt *utils.NullableTime,
	projectId *int64,
	tags []string,
	archivedAt *utils.NullableTime,
	starMarkAt *utils.NullableTime,
	givenUpAt *utils.NullableTime,
) (*UpdateTask, error) {
	updateTask := &UpdateTask{
		UserId:       userId,
		ParentTaskId: parentTaskId,
		Name:         name,
		Description:  description,
		State:        state,
		Priority:     priority,
		StartAt:      startAt,
		EndAt:        endAt,
		ProjectId:    projectId,
		Tags:         tags,
		ArchivedAt:   archivedAt,
		StarMarkAt:   starMarkAt,
		GivenUpAt:    givenUpAt,
	}
	err := updateTask.Validate()
	if err != nil {
		return nil, err
	}
	return updateTask, nil
}
