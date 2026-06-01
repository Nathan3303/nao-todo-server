package cron

import (
	"context"
	"fmt"
	"naotodoserver/application"
)

// ReminderJob 提醒扫描任务
type ReminderJob struct{}

// NewReminderJob 创建提醒扫描任务
func NewReminderJob() *ReminderJob {
	return &ReminderJob{}
}

// Run 执行提醒扫描
func (rj *ReminderJob) Run() {
	err := application.App.Task.ProcessReminders(context.TODO())
	if err != nil {
		fmt.Println("处理到期提醒失败：" + err.Error())
	}
}
