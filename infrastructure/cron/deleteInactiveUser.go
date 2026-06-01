package cron

import (
	"context"
	"fmt"
	"naotodoserver/application"
)

type DeleteDeactivedUserJob struct {
	DayOffset int8
}

func NewDeleteDeactivedUserJob(dayOffset int8) *DeleteDeactivedUserJob {
	return &DeleteDeactivedUserJob{DayOffset: dayOffset}
}

func (ddu *DeleteDeactivedUserJob) Run() {
	err := application.App.User.DeleteDeactivatedUsers(context.TODO(), ddu.DayOffset)
	if err != nil {
		fmt.Println("删除已注销用户失败：" + err.Error())
	}
}
