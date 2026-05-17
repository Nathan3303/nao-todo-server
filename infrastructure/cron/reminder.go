package cron

import (
	"fmt"
	"naotodoserver/domain/task/service"
	"naotodoserver/infrastructure/sse"
	taskRepo "naotodoserver/infrastructure/persistence/task"
	"naotodoserver/infrastructure/persistence/dbs"
	"strconv"
)

// ReminderJob 提醒扫描任务
type ReminderJob struct{}

// NewReminderJob 创建提醒扫描任务
func NewReminderJob() *ReminderJob {
	return &ReminderJob{}
}

// Run 执行提醒扫描
func (rj *ReminderJob) Run() {
	// 1. 创建仓库实例
	repo := taskRepo.NewTaskRepo(dbs.DB)
	// 2. 查询到期提醒
	tasks, err := repo.GetDueReminders(nil)
	if err != nil {
		fmt.Println("查询到期提醒失败：" + err.Error())
		return
	}
	// 3. 获取 Hub 实例
	hub := sse.GetHub()
	// 4. 处理每条到期提醒
	for _, task := range tasks {
		// 推送 SSE 事件
		hub.Publish(task.UserId, sse.ReminderEvent{
			Type:        "REMINDER",
			TaskId:      strconv.FormatInt(task.Id, 10),
			TaskName:    task.Name,
			Description: task.Description,
			RemindAt:    task.GetFormatedRemindAt(),
		})
		// 处理重复提醒
		if task.RemindRepeat != 0 {
			next := service.CalculateNextRemindAt(
				*task.RemindAt,
				task.RemindRepeat,
				task.RemindTime,
				task.RemindWeekdays,
				task.EndAt,
			)
			if next != nil {
				repo.UpdateRemindAt(nil, task.Id, next.Format("2006-01-02 15:04:05"))
			} else {
				repo.ClearRemindRepeat(nil, task.Id)
			}
		} else {
			// 单次提醒，清除 remind_at
			repo.UpdateRemindAt(nil, task.Id, "")
		}
	}
	if len(tasks) > 0 {
		fmt.Printf("已推送 %d 条到期提醒\n", len(tasks))
	}
}
