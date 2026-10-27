//go:build integration

// F-06 仓储级回归：sync upsert 的「未提供不写列」与「显式置空可生效」。
//
// 覆盖 PM 验收口径：
//
//	① 已有检查项(description=备注A) → push 不带 description ⇒ 库中仍是备注A（回退修法必红）
//	② push 带 description=备注B ⇒ 更新；显式 "" ⇒ 清空
//	③ 新建检查项不带 description ⇒ 空、不报错
//	⑤ 评论 attachments/isTopUp：未提供 ⇒ 保留；提供 ⇒ 覆盖；新建缺省 ⇒ 空/false
package task

import (
	"context"
	"fmt"
	"testing"
	"time"

	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// seedUser 插入评论创建所需用户行（CreateComment 会读取昵称/头像）
func seedUser(t *testing.T, userID int64) {
	t.Helper()
	user := &models.User{
		ModelBase: models.ModelBase{ID: userID},
		Account:   fmt.Sprintf("u%d", userID),
		Email:     fmt.Sprintf("u%d@test.local", userID),
		Password:  "x",
		Nickname:  fmt.Sprintf("nick%d", userID),
	}
	if err := testDB.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

// TestUpsertCheckItemKeepsDescriptionWhenAbsentIntegration ① ② ③
func TestUpsertCheckItemKeepsDescriptionWhenAbsentIntegration(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	const userID, taskID = int64(1061), int64(9061)
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 首次创建：显式带备注 A（走 create 路径，Id=0）
	created, upsert, err := repo.UpsertCheckItem(ctx, userID, &valueobjects.CreateTaskCheckItem{
		UserId: userID, TaskId: taskID, Name: "检查项",
		Description: "备注A", DescriptionSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("创建检查项: %v", err)
	}
	if !upsert.Created {
		t.Fatal("首次创建应 created=true")
	}

	// ① 客户端 push 不带 description（零值 + DescriptionSet=false）
	updated, _, err := repo.UpsertCheckItem(ctx, userID, &valueobjects.CreateTaskCheckItem{
		UserId: userID, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt:      now.Add(2 * time.Second),
		TaskId:         taskID,
		Name:           "检查项改名",
		IsDone:         true,
		SortId:         5,
		DescriptionSet: false,
	})
	if err != nil {
		t.Fatalf("覆盖检查项: %v", err)
	}
	if updated.Description != "备注A" {
		t.Fatalf("未提供 description 必须保留原值，实际 %q", updated.Description)
	}
	if updated.Name != "检查项改名" || !updated.IsDone || updated.SortId != 5 {
		t.Fatalf("其它字段应正常覆盖，实际 %+v", updated)
	}
	var row models.TaskCheckItem
	if err := testDB.First(&row, "id = ?", created.Id).Error; err != nil {
		t.Fatalf("读回检查项: %v", err)
	}
	if row.Description != "备注A" {
		t.Fatalf("库中 description = %q, want 备注A（回退修法必红）", row.Description)
	}

	// ② 带 description 值 ⇒ 更新
	updated, _, err = repo.UpsertCheckItem(ctx, userID, &valueobjects.CreateTaskCheckItem{
		UserId: userID, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(4 * time.Second), TaskId: taskID, Name: "检查项改名",
		Description: "备注B", DescriptionSet: true,
	})
	if err != nil {
		t.Fatalf("更新 note: %v", err)
	}
	if updated.Description != "备注B" {
		t.Fatalf("显式 description 应更新，实际 %q", updated.Description)
	}

	// ② 显式 "" ⇒ 清空（用户仍能清备注）
	updated, _, err = repo.UpsertCheckItem(ctx, userID, &valueobjects.CreateTaskCheckItem{
		UserId: userID, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(6 * time.Second), TaskId: taskID, Name: "检查项改名",
		Description: "", DescriptionSet: true,
	})
	if err != nil {
		t.Fatalf("清空 note: %v", err)
	}
	if updated.Description != "" {
		t.Fatalf("显式置空应生效，实际 %q", updated.Description)
	}

	// ③ 新建不带 description ⇒ 空、不报错
	fresh, _, err := repo.UpsertCheckItem(ctx, userID, &valueobjects.CreateTaskCheckItem{
		UserId: userID, TaskId: taskID, Name: "无备注检查项",
	})
	if err != nil {
		t.Fatalf("新建无备注检查项: %v", err)
	}
	if fresh.Description != "" {
		t.Fatalf("新建未提供 description 应为空，实际 %q", fresh.Description)
	}
}

// TestUpsertCommentKeepsAttachmentsWhenAbsentIntegration ⑤
func TestUpsertCommentKeepsAttachmentsWhenAbsentIntegration(t *testing.T) {
	cleanTasks(t)
	const userID, taskID = int64(1062), int64(9062)
	seedUser(t, userID)
	repo := NewTaskRepo(testDB)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	created, _, err := repo.UpsertComment(ctx, userID, &valueobjects.CreateTaskComment{
		UserId: userID, TaskId: taskID, Content: "评论",
		Attachments: []string{"f1"}, AttachmentsSet: true,
		IsTopUp: true, IsTopUpSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("创建评论: %v", err)
	}

	// 未提供 attachments/isTopUp ⇒ 保留原值（只改 content）
	updated, _, err := repo.UpsertComment(ctx, userID, &valueobjects.CreateTaskComment{
		UserId: userID, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(2 * time.Second), TaskId: taskID, Content: "评论改",
	})
	if err != nil {
		t.Fatalf("覆盖评论: %v", err)
	}
	if len(updated.Attachments) != 1 || updated.Attachments[0] != "f1" || !updated.IsTopUp {
		t.Fatalf("未提供的 attachments/isTopUp 必须保留原值，实际 %+v", updated)
	}
	if updated.Content != "评论改" {
		t.Fatalf("content 应正常覆盖，实际 %q", updated.Content)
	}

	// 显式提供 ⇒ 覆盖
	updated, _, err = repo.UpsertComment(ctx, userID, &valueobjects.CreateTaskComment{
		UserId: userID, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(4 * time.Second), TaskId: taskID, Content: "评论改",
		Attachments: []string{"f2", "f3"}, AttachmentsSet: true,
		IsTopUp: false, IsTopUpSet: true,
	})
	if err != nil {
		t.Fatalf("覆盖评论附件: %v", err)
	}
	if len(updated.Attachments) != 2 || updated.IsTopUp {
		t.Fatalf("显式 attachments/isTopUp 应覆盖，实际 %+v", updated)
	}

	// 新建缺省 ⇒ 空 attachments / false isTopUp
	fresh, _, err := repo.UpsertComment(ctx, userID, &valueobjects.CreateTaskComment{
		UserId: userID, TaskId: taskID, Content: "新评论",
	})
	if err != nil {
		t.Fatalf("新建评论: %v", err)
	}
	if len(fresh.Attachments) != 0 || fresh.IsTopUp {
		t.Fatalf("新建缺省应为空/false，实际 %+v", fresh)
	}
}
