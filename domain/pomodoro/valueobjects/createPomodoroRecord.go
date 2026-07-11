package valueobjects

import (
	"errors"
	"naotodoserver/domain/types"
)

// CreatePomodoroRecord 创建待办任务番茄工作记录值对象
type CreatePomodoroRecord struct {
	UserId      int64
	SessionId   string
	PomodoroId  int64
	Type        uint8
	TaskId      int64
	TaskName    string
	Description string
	StartAt     types.NullableTime
	EndAt       types.NullableTime
	Duration    uint16
	Note        string
}

// Validate 验证创建待办任务番茄工作记录值对象是否有效
func (vo *CreatePomodoroRecord) Validate() error {
	if vo.SessionId == "" {
		return errors.New("会话 ID 不能为空")
	}
	if vo.TaskId <= 0 {
		return errors.New("任务 ID 无效")
	}
	if vo.TaskName == "" {
		return errors.New("任务名称不能为空")
	}
	if vo.StartAt.IsNull {
		return errors.New("开始时间不能为空")
	}
	if vo.EndAt.IsNull {
		return errors.New("结束时间不能为空")
	}
	if vo.Duration <= 0 {
		return errors.New("专注时长必须大于 0")
	}
	if vo.EndAt.Time.Before(vo.StartAt.Time) {
		return errors.New("结束时间不能早于开始时间")
	}
	return nil
}

// NewCreatePomodoroRecord 创建待办任务番茄工作记录值对象
func NewCreatePomodoroRecord(
	userId int64,
	sessionId string,
	pomodoroId int64,
	pomodoroType uint8,
	taskId int64,
	taskName string,
	description string,
	startAt string,
	endAt string,
	duration uint16,
	note string,
) (*CreatePomodoroRecord, error) {
	vo := &CreatePomodoroRecord{
		UserId:      userId,
		SessionId:   sessionId,
		PomodoroId:  pomodoroId,
		Type:        pomodoroType,
		TaskId:      taskId,
		TaskName:    taskName,
		Description: description,
		StartAt:     types.NewNullableTimeByTimeStr(startAt),
		EndAt:       types.NewNullableTimeByTimeStr(endAt),
		Duration:    duration,
		Note:        note,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	return vo, nil
}
