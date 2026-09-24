//go:build integration

// DEF-12 ① + DP-1=(b) 服务端回归：任务 List 的归档过滤语义与 sort 白名单。
// 运行前需测试 MySQL：见 repoImpl_integration_test.go 顶部说明。
package task

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/task/valueobjects"
	"naotodoserver/infrastructure/persistence/models"
)

// seedArchivedTask 灌入一条任务，archived=true 时置 archived_at（不经 List 归档路径）
func seedArchivedTask(t *testing.T, repo *TaskRepoImpl, id, userID int64, name string, archived bool) {
	t.Helper()
	ctx := context.Background()
	if _, err := repo.Create(ctx, userID, newTaskVO(id, name, time.Now(), time.Now())); err != nil {
		t.Fatalf("灌入任务 %d 失败: %v", id, err)
	}
	if archived {
		if err := testDB.Model(&models.Task{}).Where("id = ?", id).
			UpdateColumn("archived_at", time.Now()).Error; err != nil {
			t.Fatalf("标记任务 %d 归档失败: %v", id, err)
		}
	}
}

// TestListArchivedFilterSemantics ① 显式 false ⇒ 排除归档；② 未传（零值）⇒ 同路径排除归档；
// 显式 true ⇒ 仅归档任务。
func TestListArchivedFilterSemantics(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2101
	seedArchivedTask(t, repo, 9301, userID, "活跃任务", false)
	seedArchivedTask(t, repo, 9302, userID, "已归档任务", true)

	// ① 显式 isArchived=false ⇒ archived_at IS NULL
	got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ParentTaskId: 0, IsArchived: false, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9301] || got[9302] {
		t.Fatalf("显式 false 应仅返回未归档，实际 %v", got)
	}

	// ② 未传 isArchived（Go 零值 false）⇒ 同样排除归档，与显式 false 同结果
	got = listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9301] || got[9302] {
		t.Fatalf("未传 isArchived 应默认排除归档，实际 %v", got)
	}

	// 显式 true ⇒ archived_at IS NOT NULL
	got = listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ParentTaskId: 0, IsArchived: true, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9302] || got[9301] {
		t.Fatalf("显式 true 应仅返回归档，实际 %v", got)
	}
}

// TestListArchivedWithSort ③ 合法 sort 生效且不改变归档过滤；注入样例 sort 拒绝/回落（不得 500），
// 归档仍被排除（归档态不受 sort 影响）。
func TestListArchivedWithSort(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2102
	seedArchivedTask(t, repo, 9311, userID, "b-活跃", false)
	seedArchivedTask(t, repo, 9312, userID, "a-活跃", false)
	seedArchivedTask(t, repo, 9313, userID, "a-归档", true)

	// 合法 sort：name:asc 生效（9312 先于 9311），归档 9313 仍排除
	entities, _, err := repo.List(context.Background(), userID, &valueobjects.QueryTask{
		UserId: userID, ParentTaskId: 0, Sort: "name:asc", Page: 1, Limit: 100,
	}, &valueobjects.Pagination{Page: 1, Limit: 100})
	if err != nil {
		t.Fatalf("合法 sort 列表失败: %v", err)
	}
	if len(entities) != 2 || entities[0].Id != 9312 || entities[1].Id != 9311 {
		t.Fatalf("name:asc 顺序错误或归档未排除: %+v", entities)
	}

	// 注入/非法 sort：不得报错（无 500），回落默认序，且归档仍排除
	for _, s := range []string{"id;SELECT 1:asc", "name:sideways", "name:asc:desc", "unknownField:asc"} {
		got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
			UserId: userID, ParentTaskId: 0, Sort: s, Page: 1, Limit: 100,
		})
		if len(got) != 2 || !got[9311] || !got[9312] || got[9313] {
			t.Fatalf("sort=%q 应回落默认序且仍排除归档，实际 %v", s, got)
		}
	}
}

// TestListSyncStillReturnsArchived 不变量：/sync/pull 链路（ListSync）不走 ByTaskArchived，
// 归档任务照常返回 ⇒ 客户端镜像完整（DP-1=(b) 不影响同步面）。
func TestListSyncStillReturnsArchived(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2103
	seedArchivedTask(t, repo, 9321, userID, "活跃任务", false)
	seedArchivedTask(t, repo, 9322, userID, "已归档任务", true)

	entities, err := repo.ListSync(context.Background(), userID, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync 失败: %v", err)
	}
	got := make(map[int64]bool, len(entities))
	for _, e := range entities {
		got[e.Id] = true
	}
	if len(got) != 2 || !got[9321] || !got[9322] {
		t.Fatalf("/sync/pull 应返回含归档的全部任务（镜像完整），实际 %v", got)
	}
}
