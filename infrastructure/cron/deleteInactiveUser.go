package cron

import (
	"context"
	"fmt"
	userApp "naotodoserver/application/user"
)

type DeleteDeactivedUserJob struct {
	DayOffset int8
	UserApp   userApp.UserApp
}

func NewDeleteDeactivedUserJob(dayOffset int8, userApp userApp.UserApp) *DeleteDeactivedUserJob {
	return &DeleteDeactivedUserJob{DayOffset: dayOffset, UserApp: userApp}
}

func (ddu *DeleteDeactivedUserJob) Run() {
	err := ddu.UserApp.DeleteDeactivatedUsers(context.TODO(), ddu.DayOffset)
	if err != nil {
		fmt.Println("删除已注销用户失败：" + err.Error())
	}
}
