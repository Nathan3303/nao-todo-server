//go:build integration

// T327 仓储级：标签删除的级联清理（RemoveTagFromTasks）属「派生写」，
// 登记的行版本必须与库中 updated_at 逐字相等（否则客户端 base 收敛到错值）。
package task

import (
	"context"
	"testing"
	"time"

	"naotodoserver/infrastructure/derived"
	"naotodoserver/infrastructure/persistence/models"
)

func TestDerivedRecorder_RemoveTagFromTasks(t *testing.T) {
	cleanTasks(t)
	const uid, tid, tagID = int64(424242), int64(7001), int64(7002)
	base := time.Now().UTC().Add(-time.Hour).Truncate(time.Millisecond)
	if err := testDB.Create(&models.Task{
		ModelBase: models.ModelBase{ID: tid, CreatedAt: base, UpdatedAt: base},
		UserId:    uid, Name: "标签任务", Tags: []string{"7002", "7003"},
	}).Error; err != nil {
		t.Fatalf("seed task: %v", err)
	}

	rec := derived.NewRecorder()
	ctx := derived.WithRecorder(context.Background(), rec)
	if err := NewTaskRepo(testDB).RemoveTagFromTasks(ctx, uid, tagID); err != nil {
		t.Fatalf("RemoveTagFromTasks: %v", err)
	}

	var row models.Task
	if err := testDB.Unscoped().First(&row, "id = ?", tid).Error; err != nil {
		t.Fatalf("读回任务: %v", err)
	}
	ups := rec.Updates()
	if len(ups) != 1 || ups[0].Table != derived.TableTasks || ups[0].Id != "7001" {
		t.Fatalf("应登记 1 条 tasks/7001 派生写, got %+v", ups)
	}
	if !ups[0].UpdatedAt.Equal(row.UpdatedAt) {
		t.Fatalf("回执版本 %v != 库中 %v（派生行回执必须逐字相等）",
			ups[0].UpdatedAt.UTC().Format(time.RFC3339Nano), row.UpdatedAt.UTC().Format(time.RFC3339Nano))
	}
	if len(row.Tags) != 1 || row.Tags[0] != "7003" {
		t.Fatalf("标签引用应被清理为 [7003], got %v", row.Tags)
	}
}
