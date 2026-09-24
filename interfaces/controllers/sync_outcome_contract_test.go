// T143 契约测试：/sync/push 逐条回传 outcome（additive）。
//
// 覆盖全部 outcome 分支（applied / noop / conflict / skipped / error）以及
// 「7 张表均接线」的正向断言。语义必须与 domain/types.DecideUpsert 判定同源：
// 映射只在 syncOutcomeOf / syncErrOutcome 一处发生。
package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	pomodoroApp "naotodoserver/application/pomodoro"
	pomodoroDto "naotodoserver/application/pomodoro/dto"
	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	tagApp "naotodoserver/application/tag"
	tagDto "naotodoserver/application/tag/dto"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domerr "naotodoserver/domain/errors"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// --- fake app：仅覆盖 Push 用到的 Create*；其余方法由内嵌接口占位（被调用即 panic） ---

type fakeTaskApp struct {
	taskApp.TaskApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakeTaskApp) CreateTask(
	_ context.Context, _ int64, _ *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &taskDto.GetTaskRes{Id: "9001", UpdatedAt: "2026-09-24T00:00:00Z"}, f.upsert, nil
}

type fakeCheckItemApp struct {
	taskApp.TaskCheckItemApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakeCheckItemApp) CreateTaskCheckItem(
	_ context.Context, _ int64, _ *taskDto.CreateTaskCheckItemReq,
) (*taskDto.CreateTaskCheckItemRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &taskDto.CreateTaskCheckItemRes{Id: "9002", UpdatedAt: "2026-09-24T00:00:00Z"}, f.upsert, nil
}

type fakeCommentApp struct {
	taskApp.TaskCommentApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakeCommentApp) CreateTaskComment(
	_ context.Context, _ int64, _ *taskDto.CreateTaskCommentReq,
) (*taskDto.TaskCommentRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &taskDto.TaskCommentRes{Id: "9003", UpdatedAt: "2026-09-24T00:00:00Z"}, f.upsert, nil
}

type fakeProjectApp struct {
	projectApp.ProjectApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakeProjectApp) Create(
	_ context.Context, _ int64, _ *projectDto.CreateProjectReq,
) (*projectDto.CreateProjectRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &projectDto.CreateProjectRes{Id: "9101", UpdatedAt: "2026-09-24T00:00:00Z"}, f.upsert, nil
}

type fakeTagApp struct {
	tagApp.TagApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakeTagApp) CreateTag(
	_ context.Context, _ int64, _ *tagDto.CreateTagReq,
) (*tagDto.CreateTagRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &tagDto.CreateTagRes{Id: "9201", UpdatedAt: "2026-09-24T00:00:00Z"}, f.upsert, nil
}

type fakePomodoroApp struct {
	pomodoroApp.PomodoroApp
	upsert domaintypes.UpsertResult
	err    error
}

func (f *fakePomodoroApp) CreatePomodoro(
	_ context.Context, _ int64, _ *pomodoroDto.CreatePomodoroReq,
) (*pomodoroDto.CreatePomodoroRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &pomodoroDto.CreatePomodoroRes{
		ResBase: pomodoroDto.ResBase{Id: "9301", UpdatedAt: "2026-09-24T00:00:00Z"},
	}, f.upsert, nil
}

func (f *fakePomodoroApp) Create(
	_ context.Context, _ int64, _ *pomodoroDto.CreatePomodoroRecordReq,
) (*pomodoroDto.CreatePomodoroRecordRes, domaintypes.UpsertResult, error) {
	if f.err != nil {
		return nil, domaintypes.UpsertResult{}, f.err
	}
	return &pomodoroDto.CreatePomodoroRecordRes{
		ResBase: pomodoroDto.ResBase{Id: "9401", UpdatedAt: "2026-09-24T00:00:00Z"},
	}, f.upsert, nil
}

// pushSync 以已登录用户身份调用真实 Push 控制器并解析响应
func pushSync(t *testing.T, c *SyncController, body string) types.SyncPushRes {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodPost, "/sync/push", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(iCtx.SetUserId(req.Context(), 1001))
	gc.Request = req
	c.Push(gc)
	if w.Code != http.StatusOK {
		t.Fatalf("Push HTTP 状态 = %d, body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Data types.SyncPushRes `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("解析响应失败: %v, body=%s", err, w.Body.String())
	}
	return resp.Data
}

const taskOnlyBody = `{"tasks":[{"name":"t","state":"pending","priority":"medium"}]}`

// TestSyncPushOutcome_AllBranches 逐分支断言 outcome 语义（applied / noop / conflict / error）
func TestSyncPushOutcome_AllBranches(t *testing.T) {
	cases := []struct {
		name       string
		upsert     domaintypes.UpsertResult
		err        error
		wantOut    string
		wantErrMsg bool
	}{
		{"覆盖/新建 → applied", domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite, Created: true}, nil, types.SyncOutcomeApplied, false},
		{"LWW 拒绝 → noop", domaintypes.UpsertResult{Outcome: domaintypes.UpsertNoop}, nil, types.SyncOutcomeNoop, false},
		{"ID 碰撞 → conflict", domaintypes.UpsertResult{}, domerr.ErrIDConflict, types.SyncOutcomeConflict, true},
		{"其他失败 → error", domaintypes.UpsertResult{}, errors.New("boom"), types.SyncOutcomeError, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			app := &fakeTaskApp{upsert: tc.upsert, err: tc.err}
			c := NewSyncController(app, nil, nil, nil, nil, nil)
			got := pushSync(t, c, taskOnlyBody)
			if len(got.Results) != 1 {
				t.Fatalf("results 条数 = %d, want 1", len(got.Results))
			}
			r := got.Results[0]
			if r.Outcome != tc.wantOut {
				t.Fatalf("outcome = %q, want %q", r.Outcome, tc.wantOut)
			}
			if tc.wantErrMsg && r.Error == "" {
				t.Fatal("失败分支 error 不应为空")
			}
			if !tc.wantErrMsg && r.Error != "" {
				t.Fatalf("成功分支 error = %q, want 空", r.Error)
			}
		})
	}
}

// TestSyncPushOutcome_Skipped 只追加资源删除请求 → skipped（保留既有 Skipped 字段语义）
func TestSyncPushOutcome_Skipped(t *testing.T) {
	app := &fakeTaskApp{}
	c := NewSyncController(app, nil, nil, nil, nil, nil)
	got := pushSync(t, c, `{"deletions":[{"table":"pomodoroRecords","id":"9401"}]}`)
	if len(got.Results) != 1 {
		t.Fatalf("results 条数 = %d, want 1", len(got.Results))
	}
	r := got.Results[0]
	if r.Outcome != types.SyncOutcomeSkipped {
		t.Fatalf("outcome = %q, want %q", r.Outcome, types.SyncOutcomeSkipped)
	}
	if !r.Skipped {
		t.Fatal("skipped 字段应保持 true")
	}
}

// TestSyncPushOutcome_AllTablesWired 7 张表全部接线：每表一条 applied
func TestSyncPushOutcome_AllTablesWired(t *testing.T) {
	applied := domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite, Created: true}
	c := NewSyncController(
		&fakeTaskApp{upsert: applied},
		&fakeCheckItemApp{upsert: applied},
		&fakeCommentApp{upsert: applied},
		&fakeProjectApp{upsert: applied},
		&fakeTagApp{upsert: applied},
		&fakePomodoroApp{upsert: applied},
	)
	body := `{
		"tasks":[{"name":"t","state":"pending","priority":"medium"}],
		"taskCheckItems":[{"taskId":"9001","name":"ci"}],
		"taskComments":[{"taskId":"9001","content":"c"}],
		"projects":[{"name":"p"}],
		"tags":[{"name":"g","color":"#ffffff"}],
		"pomodoros":[{"type":1,"name":"pm","duration":1500}],
		"pomodoroRecords":[{"sessionId":"s1","type":1,"startAt":"2026-09-24T00:00:00Z","endAt":"2026-09-24T00:25:00Z","duration":1500}]
	}`
	got := pushSync(t, c, body)
	if len(got.Results) != 7 {
		t.Fatalf("results 条数 = %d, want 7", len(got.Results))
	}
	wantTables := map[string]bool{
		"tasks": false, "taskCheckItems": false, "taskComments": false,
		"projects": false, "tags": false, "pomodoros": false, "pomodoroRecords": false,
	}
	for _, r := range got.Results {
		if r.Outcome != types.SyncOutcomeApplied {
			t.Fatalf("表 %s outcome = %q, want %q", r.Table, r.Outcome, types.SyncOutcomeApplied)
		}
		if _, ok := wantTables[r.Table]; !ok {
			t.Fatalf("未知表 %q", r.Table)
		}
		wantTables[r.Table] = true
	}
	for table, seen := range wantTables {
		if !seen {
			t.Fatalf("表 %s 未回传 outcome", table)
		}
	}
}

// TestSyncPushOutcome_NoOutcomeFieldRegression 既有字段语义不变：success 仍带 serverUpdatedAt，
// 失败仍带 error（additive 只新增 outcome）。
func TestSyncPushOutcome_NoOutcomeFieldRegression(t *testing.T) {
	app := &fakeTaskApp{upsert: domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite, Created: true}}
	c := NewSyncController(app, nil, nil, nil, nil, nil)
	got := pushSync(t, c, taskOnlyBody)
	if got.Results[0].ServerUpdatedAt == "" {
		t.Fatal("serverUpdatedAt 不应为空（既有字段语义不得破坏）")
	}
	if got.ServerTime == "" {
		t.Fatal("serverTime 不应为空（既有字段语义不得破坏）")
	}
}
