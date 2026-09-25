//go:build integration

// Copy 校验失败的应用层/持久化行为（T4 回归）：校验错误必须作为 error 冒泡并回滚，不得静默无效或 panic。
package task

import (
	"context"
	"strings"
	"testing"

	"naotodoserver/application/idutil"
	"naotodoserver/infrastructure/persistence/models"
)

// TestCopy_ValidateFailure_AppLayerGivenSourceNameAtLimit 源任务名称已达 256 上限时，
// 复制品名称超限 ⇒ CopyTask 返回非 nil error、事务回滚、不产生任何新行。
func TestCopy_ValidateFailure_AppLayerGivenSourceNameAtLimit(t *testing.T) {
	cleanTasks(t)
	s := newFullStack(t)
	insertProject(t, testProjA)

	src := createTask(t, s, taskReq(strings.Repeat("任", 256), testProjA, 0, nil))

	var before int64
	if err := testDB.Model(&models.Task{}).Count(&before).Error; err != nil {
		t.Fatalf("统计任务行: %v", err)
	}

	_, err := s.taskApp.CopyTask(context.Background(), testUserID, idutil.FormatID(src))

	if err == nil {
		t.Fatal("CopyTask 校验失败必须返回非 nil error（不得吞错导致上层继续使用 nil 实体）")
	}
	if !strings.Contains(err.Error(), "任务名称最多256个字符") {
		t.Fatalf("CopyTask 返回了非预期错误：%v", err)
	}
	var after int64
	if err := testDB.Model(&models.Task{}).Count(&after).Error; err != nil {
		t.Fatalf("统计任务行: %v", err)
	}
	if after != before {
		t.Fatalf("校验失败不得写入：任务行数 before %d, after %d", before, after)
	}
}
