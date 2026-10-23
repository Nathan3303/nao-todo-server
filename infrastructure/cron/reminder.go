package cron

import (
	"context"
	"fmt"

	taskApp "naotodoserver/application/task"
)

// ReminderJob 提醒扫描任务
type ReminderJob struct {
	TaskApp taskApp.TaskApp
	// Enabled 是否改写提醒字段（对应配置 reminder.cronFieldWriteBack）。
	// false = 停用（默认）：跳过 ProcessReminders，不再清空一次性 remind_at、不再推进重复提醒。
	Enabled bool
}

// NewReminderJob 创建提醒扫描任务
// enabled 对应配置 reminder.cronFieldWriteBack，默认 false（停用字段改写，提醒判定下移到客户端）。
func NewReminderJob(taskApp taskApp.TaskApp, enabled bool) *ReminderJob {
	return &ReminderJob{TaskApp: taskApp, Enabled: enabled}
}

// Run 执行提醒扫描
func (rj *ReminderJob) Run() {
	if !rj.Enabled {
		// 停用期间保留代码路径，仅跳过执行：字段不改写 ⇒ 到期痕迹留存，客户端可补弹；
		// 同时避免对同一条到期任务每分钟重复推送提醒事件。
		return
	}
	err := rj.TaskApp.ProcessReminders(context.TODO())
	if err != nil {
		fmt.Println("处理到期提醒失败：" + err.Error())
	}
}
