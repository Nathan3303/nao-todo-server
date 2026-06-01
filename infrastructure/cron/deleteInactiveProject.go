package cron

import (
	"context"
	"fmt"
	"naotodoserver/application"
)

type DeleteDeactivedProjectJob struct {
	DayOffset int8
}

func NewDeleteDeactivedProjectJob(dayOffset int8) *DeleteDeactivedProjectJob {
	return &DeleteDeactivedProjectJob{DayOffset: dayOffset}
}

func (ddp *DeleteDeactivedProjectJob) Run() {
	err := application.App.Project.DeleteDeactivatedProjects(context.TODO(), ddp.DayOffset)
	if err != nil {
		fmt.Println("删除已注销项目失败：" + err.Error())
	}
}
