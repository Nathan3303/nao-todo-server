//go:build integration

// F-06 仓储级回归（projects.sortId）：未提供 ⇒ 保留库中原值；显式 0/非 0 ⇒ 写入并可 pull 读回。
package project

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	"naotodoserver/domain/project/valueobjects"
	"naotodoserver/infrastructure/persistence/cache"
)

const sortIdTestUser = int64(3011)

// cleanSortIdUser 仅清理本用例用户，避免与其它包集成测试互删
func cleanSortIdUser(t *testing.T) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM project_preferences WHERE user_id = ?",
		"DELETE FROM projects WHERE user_id = ?",
	} {
		if err := testDB.Exec(stmt, sortIdTestUser).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

func TestUpsertProjectKeepsSortIdWhenAbsentIntegration(t *testing.T) {
	cleanSortIdUser(t)
	addr := os.Getenv("NAO_TEST_REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	repo := NewProjectRepo(testDB, cache.NewCache(redis.NewClient(&redis.Options{Addr: addr})))
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	// 首次创建：显式 sortId=300
	created, _, err := repo.Upsert(ctx, sortIdTestUser, &valueobjects.CreateProject{
		UserId: sortIdTestUser, Name: "清单",
		SortId: 300, SortIdSet: true,
		CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("创建清单: %v", err)
	}
	if created.SortId != 300 {
		t.Fatalf("创建后 sortId = %d, want 300", created.SortId)
	}

	// 未提供 sortId ⇒ 保留 300（修复「每次 push 归零」），其它字段仍覆盖
	updated, _, err := repo.Upsert(ctx, sortIdTestUser, &valueobjects.CreateProject{
		UserId: sortIdTestUser, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(2 * time.Second), Name: "清单改名",
	})
	if err != nil {
		t.Fatalf("覆盖清单: %v", err)
	}
	if updated.SortId != 300 {
		t.Fatalf("未提供 sortId 必须保留 300，实际 %d", updated.SortId)
	}
	if updated.Name != "清单改名" {
		t.Fatalf("name 应正常覆盖，实际 %q", updated.Name)
	}

	// pull 读回：ListSync 返回库中 sortId
	list, err := repo.ListSync(ctx, sortIdTestUser, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if len(list) != 1 || list[0].SortId != 300 {
		t.Fatalf("pull 读回 sortId 错误: %+v", list)
	}

	// 显式 0 ⇒ 写入 0（0 是合法排序值）
	updated, _, err = repo.Upsert(ctx, sortIdTestUser, &valueobjects.CreateProject{
		UserId: sortIdTestUser, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(4 * time.Second), Name: "清单改名",
		SortId: 0, SortIdSet: true,
	})
	if err != nil {
		t.Fatalf("显式 0 覆盖: %v", err)
	}
	if updated.SortId != 0 {
		t.Fatalf("显式 0 应写入，实际 %d", updated.SortId)
	}

	// 显式 700 ⇒ 写入并可 pull 读回
	if _, _, err = repo.Upsert(ctx, sortIdTestUser, &valueobjects.CreateProject{
		UserId: sortIdTestUser, Id: created.Id, CreatedAt: created.CreatedAt,
		UpdatedAt: now.Add(6 * time.Second), Name: "清单改名",
		SortId: 700, SortIdSet: true,
	}); err != nil {
		t.Fatalf("显式 700 覆盖: %v", err)
	}
	list, err = repo.ListSync(ctx, sortIdTestUser, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if len(list) != 1 || list[0].SortId != 700 {
		t.Fatalf("pull 读回 sortId=700 失败: %+v", list)
	}
}
