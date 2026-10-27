package controllers

import (
	"strings"
	"testing"

	"naotodoserver/interfaces/types"
)

// TestValidateSyncPushBatch 单表单次条数上限（SEC-01）：
// 等于上限放行（向后兼容），超限拒绝并给出表名与实际条数。
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
		for _, want := range []string{"tasks", "500", "501"} {
			if !strings.Contains(err.Error(), want) {
				t.Fatalf("错误信息缺少 %q，实际: %v", want, err)
			}
		}
	})

	t.Run("deletions 超限拒绝", func(t *testing.T) {
		req := &types.SyncPushReq{Deletions: make([]types.SyncDeletion, maxSyncPushBatch+1)}
		err := validateSyncPushBatch(req)
		if err == nil || !strings.Contains(err.Error(), "deletions") {
			t.Fatalf("501 条 deletions 必须被拒，实际: %v", err)
		}
	})

	t.Run("projects/tags 超限拒绝", func(t *testing.T) {
		for name, req := range map[string]*types.SyncPushReq{
			"projects": {Projects: make([]types.SyncProjectPushItem, maxSyncPushBatch+1)},
			"tags":     {Tags: make([]types.SyncTagPushItem, maxSyncPushBatch+1)},
		} {
			err := validateSyncPushBatch(req)
			if err == nil || !strings.Contains(err.Error(), name) {
				t.Fatalf("%s 超限必须被拒，实际: %v", name, err)
			}
		}
	})
}
