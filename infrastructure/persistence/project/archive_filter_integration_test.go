//go:build integration

// DP-1=(b) 服务端回归：清单列表默认排除归档（未传/显式 false 同路径），
// /sync/pull 与按 id 单取、取消归档路径不受影响（镜像完整）。
// 运行前需测试 MySQL：见 preferenceRepoImpl_integration_test.go 顶部说明。
package project

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"

	"naotodoserver/domain/project/entities"
	"naotodoserver/domain/project/repositories"
	"naotodoserver/domain/types"
	"naotodoserver/infrastructure/persistence/cache"
	"naotodoserver/infrastructure/persistence/models"
)

// projectArchiveFilterTestUser 本用例专用用户 ID（避开其它包的测试用户区间）
const projectArchiveFilterTestUser = 3020

// cleanProjectArchiveFilterData 清理本用例用户的清单与偏好（偏好先删，满足外键）
func cleanProjectArchiveFilterData(t *testing.T, userID int64) {
	t.Helper()
	for _, stmt := range []string{
		"DELETE FROM project_preferences WHERE project_id IN (SELECT id FROM projects WHERE user_id = ?)",
		"DELETE FROM projects WHERE user_id = ?",
	} {
		if err := testDB.Exec(stmt, userID).Error; err != nil {
			t.Fatalf("清理失败 %s: %v", stmt, err)
		}
	}
}

// newProjectRepoNoCache 清单 repo，缓存后端指向不可达 Redis：cache.Get/Set/Del 静默降级，
// 既不依赖测试 Redis，又保证每次都真实读库（归档过滤为 SQL 层断言）。
func newProjectRepoNoCache() repositories.Project {
	return NewProjectRepo(testDB, cache.NewCache(redis.NewClient(&redis.Options{Addr: "127.0.0.1:1"})))
}

// seedProject 灌入一条清单，archived=true 时置 archived_at（不经归档接口）
func seedProject(t *testing.T, userID, id int64, name string, archived bool) {
	t.Helper()
	if err := testDB.Create(&models.Project{
		ModelBase: models.ModelBase{ID: id},
		UserId:    userID,
		Name:      name,
	}).Error; err != nil {
		t.Fatalf("灌入清单 %d 失败: %v", id, err)
	}
	if archived {
		if err := testDB.Model(&models.Project{}).
			Where("id = ? AND user_id = ?", id, userID).
			UpdateColumn("archived_at", time.Now()).Error; err != nil {
			t.Fatalf("标记清单 %d 归档失败: %v", id, err)
		}
	}
}

func projectIDs(es []*entities.Project) map[int64]bool {
	got := make(map[int64]bool, len(es))
	for _, e := range es {
		got[e.Id] = true
	}
	return got
}

// TestProjectListArchivedFilterSemantics ① 显式 false ⇒ 排除归档；② 未传（零值）⇒ 同路径排除归档；
// ③ 显式 true ⇒ 仅归档清单。
func TestProjectListArchivedFilterSemantics(t *testing.T) {
	const userID = projectArchiveFilterTestUser
	cleanProjectArchiveFilterData(t, userID)
	t.Cleanup(func() { cleanProjectArchiveFilterData(t, userID) })
	repo := newProjectRepoNoCache()
	ctx := context.Background()
	seedProject(t, userID, 8801, "活跃清单", false)
	seedProject(t, userID, 8802, "已归档清单", true)

	// ① 显式 isArchived=false ⇒ archived_at IS NULL
	got, err := repo.GetByUserId(ctx, userID, false)
	if err != nil {
		t.Fatalf("GetByUserId(false): %v", err)
	}
	if ids := projectIDs(got); len(ids) != 1 || !ids[8801] || ids[8802] {
		t.Fatalf("显式 false 应仅返回未归档清单，实际 %v", ids)
	}

	// ② 未传 isArchived（Go 零值 false）⇒ 同样排除归档，与显式 false 同结果
	var unset bool
	gotUnset, err := repo.GetByUserId(ctx, userID, unset)
	if err != nil {
		t.Fatalf("GetByUserId(未传): %v", err)
	}
	if ids := projectIDs(gotUnset); len(ids) != 1 || !ids[8801] || ids[8802] {
		t.Fatalf("未传 isArchived 应默认排除归档，实际 %v", ids)
	}

	// ③ 显式 isArchived=true ⇒ archived_at IS NOT NULL
	gotArchived, err := repo.GetByUserId(ctx, userID, true)
	if err != nil {
		t.Fatalf("GetByUserId(true): %v", err)
	}
	if ids := projectIDs(gotArchived); len(ids) != 1 || !ids[8802] || ids[8801] {
		t.Fatalf("显式 true 应仅返回归档清单，实际 %v", ids)
	}
}

// TestProjectArchivedStillReachableBySyncIdAndUnarchive 不变量：其余归档入口不受误伤 ——
// ① /sync/pull（ListSync）照常返回归档清单（镜像完整，与任务面 TestListSyncStillReturnsArchived 对称）；
// ② 按 id 单取（GetById）仍可取到已归档清单；
// ③ 取消归档（UpdateState 置空 archived_at）后默认列表恢复可见。
func TestProjectArchivedStillReachableBySyncIdAndUnarchive(t *testing.T) {
	const userID = projectArchiveFilterTestUser
	cleanProjectArchiveFilterData(t, userID)
	t.Cleanup(func() { cleanProjectArchiveFilterData(t, userID) })
	repo := newProjectRepoNoCache()
	ctx := context.Background()
	seedProject(t, userID, 8811, "活跃清单", false)
	seedProject(t, userID, 8812, "已归档清单", true)

	// ① /sync/pull 链路：含归档（镜像完整）
	synced, err := repo.ListSync(ctx, userID, time.Time{}, 0, 100)
	if err != nil {
		t.Fatalf("ListSync: %v", err)
	}
	if ids := projectIDs(synced); len(ids) != 2 || !ids[8811] || !ids[8812] {
		t.Fatalf("/sync/pull 应返回含归档的全部清单（镜像完整），实际 %v", ids)
	}

	// ② 按 id 单取：已归档清单仍可读
	one, err := repo.GetById(ctx, userID, 8812)
	if err != nil {
		t.Fatalf("GetById(归档清单): %v", err)
	}
	if one == nil || one.Id != 8812 || !one.ArchivedAt.Valid || one.ArchivedAt.IsNull {
		t.Fatalf("按 id 单取应仍返回已归档清单，实际 %+v", one)
	}

	// ③ 取消归档路径：置空 archived_at 后默认列表恢复可见
	if err := repo.UpdateState(
		ctx, userID, 8812,
		types.NewNullableTimeSetToNull(),
		types.NewNullableTimeNull(),
	); err != nil {
		t.Fatalf("UpdateState(取消归档): %v", err)
	}
	after, err := repo.GetByUserId(ctx, userID, false)
	if err != nil {
		t.Fatalf("GetByUserId(取消归档后): %v", err)
	}
	if ids := projectIDs(after); len(ids) != 2 || !ids[8811] || !ids[8812] {
		t.Fatalf("取消归档后默认列表应恢复可见，实际 %v", ids)
	}
}
