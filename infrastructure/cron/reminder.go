package cron

import (
	"context"
	"fmt"
	taskApp "naotodoserver/application/task"
)

// ReminderJob 提醒扫描任务
type ReminderJob struct {
	TaskApp taskApp.TaskApp
}

// NewReminderJob 创建提醒扫描任务
func NewReminderJob(taskApp taskApp.TaskApp) *ReminderJob {
	return &ReminderJob{TaskApp: taskApp}
}

// Run 执行提醒扫描
func (rj *ReminderJob) Run() {
	err := rj.TaskApp.ProcessReminders(context.TODO())
	if err != nil {
		fmt.Println("处理到期提醒失败：" + err.Error())
	}
}
