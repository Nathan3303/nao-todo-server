package cron

import (
	"context"
	"fmt"
	projectApp "naotodoserver/application/project"
)

type DeleteDeactivedProjectJob struct {
	DayOffset   int8
	ProjectApp  projectApp.ProjectApp
}

func NewDeleteDeactivedProjectJob(dayOffset int8, projectApp projectApp.ProjectApp) *DeleteDeactivedProjectJob {
	return &DeleteDeactivedProjectJob{DayOffset: dayOffset, ProjectApp: projectApp}
}

func (ddp *DeleteDeactivedProjectJob) Run() {
	err := ddp.ProjectApp.DeleteDeactivatedProjects(context.TODO(), ddp.DayOffset)
	if err != nil {
		fmt.Println("删除已注销项目失败：" + err.Error())
	}
}
