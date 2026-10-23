// AC1/AC2 回归：提醒 cron 的字段改写开关——停用时跳过扫描（不触发 ProcessReminders，
// 因而不会清空一次性 remind_at / 推进重复提醒），打开后恢复既有行为。
package cron

import (
	"context"
	"testing"

	taskApp "naotodoserver/application/task"
)

// fakeReminderTaskApp 只统计 ProcessReminders 调用次数；其余接口由内嵌 nil 占位。
type fakeReminderTaskApp struct {
	taskApp.TaskApp
	calls int
}

func (f *fakeReminderTaskApp) ProcessReminders(_ context.Context) error {
	f.calls++
	return nil
}

func TestReminderJob_WriteBackSwitch(t *testing.T) {
	// 默认（停用）：执行被跳过 ⇒ 无任何字段改写。
	disabled := &fakeReminderTaskApp{}
	NewReminderJob(disabled, false).Run()
	if disabled.calls != 0 {
		t.Fatalf("开关停用时 ProcessReminders 调用次数 = %d, want 0", disabled.calls)
	}

	// 回滚（启用）：恢复旧的每分钟扫描与字段改写。
	enabled := &fakeReminderTaskApp{}
	NewReminderJob(enabled, true).Run()
	if enabled.calls != 1 {
		t.Fatalf("开关启用时 ProcessReminders 调用次数 = %d, want 1", enabled.calls)
	}
}
