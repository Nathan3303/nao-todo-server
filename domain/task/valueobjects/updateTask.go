package valueobjects

import (
	"database/sql"
	"errors"
)

// UpdateTask 更新任务值对象
type UpdateTask struct {
	UserId       int64
	ParentTaskId *int64
	Name         *string
	Description  *string
	State        *int8
	Priority     *int8
	StartAt      sql.NullTime
	EndAt        sql.NullTime
	ProjectId    *int64
	Tags         []string
	ArchivedAt   sql.NullTime
	StarMarkAt   sql.NullTime
	GivenUpAt    sql.NullTime
}

// Validate 验证更新任务值对象
// @return error 错误信息
func (updateTask *UpdateTask) Validate() error {
	if updateTask.Name != nil && len(*updateTask.Name) == 0 {
		return errors.New("任务名称不能为空")
	}
	if updateTask.Name != nil && len(*updateTask.Name) > 256 {
		return errors.New("任务名称最多256个字符")
	}
	if updateTask.Description != nil && len(*updateTask.Description) > 512 {
		return errors.New("任务描述最多512个字符")
	}
	if updateTask.EndAt.Valid && updateTask.StartAt.Valid &&
		updateTask.EndAt.Time.Before(updateTask.StartAt.Time) {
		return errors.New("结束时间不能早于开始时间")
	}
	return nil
}

// NewUpdateTask 创建更新任务值对象
// @param userId 用户ID
// @param parentTaskId 父任务ID
// @param name 任务名称
// @param description 任务描述
// @param state 任务状态
// @param priority 任务优先级
// @param startAt 开始时间
// @param endAt 结束时间
// @param projectId 项目ID
// @param tags 标签
// @param archivedAt 归档时间
// @param starMarkAt 收藏时间
// @param givenUpAt 已完成时间
// @return 更新任务值对象
// @return error 错误信息
func NewUpdateTask(
	userId int64,
	parentTaskId *int64,
	name *string,
	description *string,
	state *int8,
	priority *int8,
	startAt sql.NullTime,
	endAt sql.NullTime,
	projectId *int64,
	tags []string,
	archivedAt sql.NullTime,
	starMarkAt sql.NullTime,
	givenUpAt sql.NullTime,
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
