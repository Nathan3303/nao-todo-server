//go:build integration

// 多值 project/tag 筛选 + relativeDate=month 的服务端过滤契约（日历下沉）。
// 运行前需测试 MySQL：见 repoImpl_integration_test.go 顶部说明。
package task

import (
	"context"
	"testing"
	"time"

	"naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

// filterTaskVO 构造带 project/tags/endAt 的任务 VO（绕过校验直设公开字段）
func filterTaskVO(id int64, name string, projectID int64, tags []string, endAt time.Time) *valueobjects.CreateTask {
	vo := newTaskVO(id, name, time.Now(), time.Now())
	vo.ProjectId = domaintypes.ProjectID(projectID)
	vo.Tags = tags
	if !endAt.IsZero() {
		vo.EndAt = domaintypes.NewNullableTimeByTime(endAt)
	}
	return vo
}

// seedFilterTasks 灌入任务并断言成功
func seedFilterTasks(t *testing.T, repo *TaskRepoImpl, userID int64, vos ...*valueobjects.CreateTask) {
	t.Helper()
	ctx := context.Background()
	for _, vo := range vos {
		if _, err := repo.Create(ctx, userID, vo); err != nil {
			t.Fatalf("灌入任务 %d 失败: %v", vo.Id, err)
		}
	}
}

// listTaskIDs 按 QueryTask 过滤并返回命中的任务 id 集合
func listTaskIDs(t *testing.T, repo *TaskRepoImpl, userID int64, q *valueobjects.QueryTask) map[int64]bool {
	t.Helper()
	ctx := context.Background()
	entities, _, err := repo.List(ctx, userID, q, &valueobjects.Pagination{Page: 1, Limit: 100})
	if err != nil {
		t.Fatalf("List 失败: %v", err)
	}
	got := make(map[int64]bool, len(entities))
	for _, e := range entities {
		got[e.Id] = true
	}
	return got
}

func TestListMultiProjectFilter(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2001
	seedFilterTasks(t, repo, userID,
		filterTaskVO(9201, "P11-A", 11, nil, time.Time{}),
		filterTaskVO(9202, "P22-B", 22, nil, time.Time{}),
		filterTaskVO(9203, "P33-C", 33, nil, time.Time{}),
	)

	// 多 project：组内 OR
	got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ProjectIds: []int64{11, 33}, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 2 || !got[9201] || !got[9203] || got[9202] {
		t.Fatalf("多 project OR 命中错误: %v", got)
	}

	// 单值语义不变：仅 P11
	got = listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ProjectIds: []int64{11}, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9201] {
		t.Fatalf("单 project 过滤错误: %v", got)
	}
}

func TestListMultiTagFilter(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2002
	seedFilterTasks(t, repo, userID,
		filterTaskVO(9211, "T1", 1, []string{"t1"}, time.Time{}),
		filterTaskVO(9212, "T2T3", 1, []string{"t2", "t3"}, time.Time{}),
		filterTaskVO(9213, "none", 1, nil, time.Time{}),
		filterTaskVO(9214, "T1T4", 1, []string{"t1", "t4"}, time.Time{}),
	)

	// 多 tag：组内 OR（tags 命中其一）
	got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, TagIds: []string{"t1", "t2"}, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 3 || !got[9211] || !got[9212] || !got[9214] || got[9213] {
		t.Fatalf("多 tag OR 命中错误: %v", got)
	}

	// 单值语义不变
	got = listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, TagIds: []string{"t3"}, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9212] {
		t.Fatalf("单 tag 过滤错误: %v", got)
	}
}

func TestListProjectAndTagAreAnded(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2003
	seedFilterTasks(t, repo, userID,
		filterTaskVO(9221, "P11-T1", 11, []string{"t1"}, time.Time{}),
		filterTaskVO(9222, "P11-T2", 11, []string{"t2"}, time.Time{}),
		filterTaskVO(9223, "P22-T1", 22, []string{"t1"}, time.Time{}),
	)

	// project 与 tag 同传：组间 AND（旧行为下 tag 会被忽略 → 会误返回 9222）
	got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ProjectIds: []int64{11}, TagIds: []string{"t1"}, ParentTaskId: 0, Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9221] {
		t.Fatalf("project×tag 组间 AND 错误: %v", got)
	}
}

func TestListRelativeDateMonth(t *testing.T) {
	cleanTasks(t)
	repo := NewTaskRepo(testDB)
	const userID = 2004

	now := time.Now()
	y, m, _ := now.Date()
	midMonth := time.Date(y, m, 15, 12, 0, 0, 0, time.Local) // 本月 15 日：窗口内
	prevMonth := midMonth.AddDate(-2, 0, 0)                  // 两月前：窗口外
	nextMonth := midMonth.AddDate(2, 0, 0)                   // 两月后：窗口外
	seedFilterTasks(t, repo, userID,
		filterTaskVO(9231, "本月到期", 1, nil, midMonth),
		filterTaskVO(9232, "两月前", 1, nil, prevMonth),
		filterTaskVO(9233, "两月后", 1, nil, nextMonth),
	)

	got := listTaskIDs(t, repo, userID, &valueobjects.QueryTask{
		UserId: userID, ParentTaskId: 0, RelativeDate: "month", Page: 1, Limit: 100,
	})
	if len(got) != 1 || !got[9231] {
		t.Fatalf("relativeDate=month 应只命中本月到期任务，实际 %v", got)
	}
}
