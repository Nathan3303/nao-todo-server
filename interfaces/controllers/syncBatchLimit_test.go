package controllers

import (
	"strconv"
	"strings"
	"testing"

	domerr "naotodoserver/domain/errors"
	"naotodoserver/interfaces/types"
)

// TestValidateSyncPushBatch 单表单次条数上限（SEC-01）：
// 等于上限放行（向后兼容），超限拒绝并给出上限值；
// 文案不回显表名（内部 schema 名，SEC-12）。
func TestValidateSyncPushBatch(t *testing.T) {
	t.Run("空请求放行", func(t *testing.T) {
		if err := validateSyncPushBatch(&types.SyncPushReq{}); err != nil {
			t.Fatalf("空请求不应被拒，实际: %v", err)
		}
	})

	t.Run("等于上限放行", func(t *testing.T) {
		req := &types.SyncPushReq{
			Tasks:     make([]types.SyncTaskPushItem, maxSyncPushBatch),
			Deletions: make([]types.SyncDeletion, maxSyncPushBatch),
		}
		if err := validateSyncPushBatch(req); err != nil {
			t.Fatalf("%d 条不应被拒，实际: %v", maxSyncPushBatch, err)
		}
	})

	t.Run("tasks 超限拒绝", func(t *testing.T) {
		req := &types.SyncPushReq{Tasks: make([]types.SyncTaskPushItem, maxSyncPushBatch+1)}
		err := validateSyncPushBatch(req)
		if err == nil {
			t.Fatal("501 条 tasks 必须被拒")
		}
		assertBatchLimitText(t, err)
	})

	t.Run("deletions 超限拒绝", func(t *testing.T) {
		req := &types.SyncPushReq{Deletions: make([]types.SyncDeletion, maxSyncPushBatch+1)}
		err := validateSyncPushBatch(req)
		if err == nil {
			t.Fatalf("501 条 deletions 必须被拒")
		}
		assertBatchLimitText(t, err)
	})

	t.Run("projects/tags 超限拒绝", func(t *testing.T) {
		for name, req := range map[string]*types.SyncPushReq{
			"projects": {Projects: make([]types.SyncProjectPushItem, maxSyncPushBatch+1)},
			"tags":     {Tags: make([]types.SyncTagPushItem, maxSyncPushBatch+1)},
		} {
			err := validateSyncPushBatch(req)
			if err == nil {
				t.Fatalf("%s 超限必须被拒", name)
			}
			assertBatchLimitText(t, err)
		}
	})
}

// assertBatchLimitText 断言批次上限文案：含上限值、不含表名/内部 schema 名（SEC-12）
func assertBatchLimitText(t *testing.T, err error) {
	t.Helper()
	text := err.Error()
	if !strings.Contains(text, strconv.Itoa(maxSyncPushBatch)) {
		t.Fatalf("文案应包含上限 %d，实际: %q", maxSyncPushBatch, text)
	}
	for _, leak := range []string{
		"tasks", "taskCheckItems", "taskComments", "projects", "tags",
		"pomodoros", "pomodoroRecords", "deletions",
	} {
		if strings.Contains(text, leak) {
			t.Fatalf("文案不得包含表名 %q，实际: %q", leak, text)
		}
	}
	if !domerr.IsPublicError(err) {
		t.Fatalf("批次上限错误应标记为可回显业务错误，实际: %T", err)
	}
}
