//go:build integration

// F-06 仓储级回归（tags.sortId）：未提供 ⇒ 保留库中原值；显式 0/非 0 ⇒ 写入并可 pull 读回。
package tag

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/tag/valueobjects"
)

func TestUpsertTagKeepsSortIdWhenAbsentIntegration(t *testing.T) {
	cleanTags(t)
	ctx := context.Background()
	const uid = int64(3012)
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 首次创建：显式 sortId=400
	created, _, err := testRepo.Upsert(ctx, uid, &valueobjects.CreateTag{
		Name: "标签", Color: "#fff",
		SortId: 400, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("创建标签: %v", err)
	}
	if created.SortId != 400 {
		t.Fatalf("创建后 sortId = %d, want 400", created.SortId)
	}

	// 未提供 sortId ⇒ 保留 400，其它字段仍覆盖
	updated, _, err := testRepo.Upsert(ctx, uid, &valueobjects.CreateTag{
		Name: "标签改名", Color: "#000",
		Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatalf("覆盖标签: %v", err)
	}
	if updated.SortId != 400 {
		t.Fatalf("未提供 sortId 必须保留 400，实际 %d", updated.SortId)
	}
	if updated.Name != "标签改名" || updated.Color != "#000" {
		t.Fatalf("其它字段应正常覆盖，实际 %+v", updated)
	}

	// pull 读回
	list, err := testRepo.ListSync(ctx, uid, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if len(list) != 1 || list[0].SortId != 400 {
		t.Fatalf("pull 读回 sortId 错误: %+v", list)
	}

	// 显式 0 ⇒ 写入 0
	updated, _, err = testRepo.Upsert(ctx, uid, &valueobjects.CreateTag{
		Name: "标签改名", Color: "#000",
		Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(4 * time.Second),
		SortId:    0, SortIdSet: true,
	})
	if err != nil {
		t.Fatalf("显式 0 覆盖: %v", err)
	}
	if updated.SortId != 0 {
		t.Fatalf("显式 0 应写入，实际 %d", updated.SortId)
	}

	// 显式 800 ⇒ 写入并可 pull 读回
	if _, _, err = testRepo.Upsert(ctx, uid, &valueobjects.CreateTag{
		Name: "标签改名", Color: "#000",
		Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(6 * time.Second),
		SortId:    800, SortIdSet: true,
	}); err != nil {
		t.Fatalf("显式 800 覆盖: %v", err)
	}
	list, err = testRepo.ListSync(ctx, uid, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if len(list) != 1 || list[0].SortId != 800 {
		t.Fatalf("pull 读回 sortId=800 失败: %+v", list)
	}
}
