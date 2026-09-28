// T327 单元契约：/sync/push 回执新增 `derivedUpdates` 为 **additive** 字段
// （为空时省略 ⇒ 旧客户端无感），且控制器把派生写收集器的记录去重、格式化为 RFC3339Milli。
package controllers

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"naotodoserver/application/idutil"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	"naotodoserver/infrastructure/derived"
	"naotodoserver/interfaces/types"
)

// TestSyncPushRes_DerivedUpdatesOmittedWhenEmpty additive 向后兼容：为空不得出现在 JSON。
func TestSyncPushRes_DerivedUpdatesOmittedWhenEmpty(t *testing.T) {
	raw, err := json.Marshal(types.SyncPushRes{
		Results:    []types.SyncResult{{Table: "tasks", Id: "1", Outcome: types.SyncOutcomeApplied}},
		ServerTime: "1756800000000",
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "derivedUpdates") {
		t.Fatalf("空派生段应省略（旧客户端无感），实际 JSON = %s", raw)
	}
	raw, err = json.Marshal(types.SyncPushRes{
		Results:        []types.SyncResult{},
		ServerTime:     "1",
		DerivedUpdates: []types.DerivedUpdate{{Table: "tasks", Id: "2", UpdatedAt: "2026-09-28T10:00:00.123Z"}},
	})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), `"derivedUpdates"`) {
		t.Fatalf("非空派生段必须出现，实际 JSON = %s", raw)
	}
}

// derivedRecEntry 脚本化派生写（id 为 int64，调用 derived.Record）。
type derivedRecEntry struct {
	table string
	id    int64
	at    time.Time
}

// derivedRecordingTaskApp 每次 CreateTask 按脚本记录派生写，用于验证控制器接线与去重。
type derivedRecordingTaskApp struct {
	taskApp.TaskApp
	script [][]derivedRecEntry
	calls  int
}

func (f *derivedRecordingTaskApp) CreateTask(
	ctx context.Context, _ int64, _ *taskDto.CreateTaskReq,
) (*taskDto.GetTaskRes, domaintypes.UpsertResult, error) {
	if f.calls < len(f.script) {
		for _, e := range f.script[f.calls] {
			derived.Record(ctx, e.table, e.id, e.at)
		}
	}
	f.calls++
	return &taskDto.GetTaskRes{Id: "9001", UpdatedAt: "2026-09-24T00:00:00Z"},
		domaintypes.UpsertResult{Outcome: domaintypes.UpsertOverwrite, Created: true}, nil
}

// TestSyncPush_DerivedUpdatesWiredDeduped 控制器把多次派生写去重为最终版本，并格式化 RFC3339Milli。
func TestSyncPush_DerivedUpdatesWiredDeduped(t *testing.T) {
	t1 := time.Date(2026, 9, 28, 10, 0, 0, 100_000_000, time.UTC)
	t2 := t1.Add(7 * time.Millisecond)
	t3 := t2.Add(3 * time.Millisecond)
	app := &derivedRecordingTaskApp{script: [][]derivedRecEntry{
		// 第一次 push：同一父行两次派生写
		{{derived.TableTasks, 42, t1}, {derived.TableTasks, 42, t2}},
		// 第二次 push：同父行再写 + 一个项目行
		{{derived.TableTasks, 42, t3}, {derived.TableProjects, 7, t3}},
	}}
	c := NewSyncController(app, nil, nil, nil, nil, nil)
	got := pushSync(t, c, `{"tasks":[{"name":"a","state":"pending","priority":"medium"},`+
		`{"name":"b","state":"pending","priority":"medium"}]}`)

	if len(got.DerivedUpdates) != 2 {
		t.Fatalf("派生段条数 = %d, want 2（同父行去重）: %+v", len(got.DerivedUpdates), got.DerivedUpdates)
	}
	if d := got.DerivedUpdates[0]; d.Table != derived.TableTasks || d.Id != "42" ||
		d.UpdatedAt != idutil.FormatTimeMilli(t3) {
		t.Fatalf("tasks/42 应为最终版本 %s, got %+v", idutil.FormatTimeMilli(t3), d)
	}
	if d := got.DerivedUpdates[1]; d.Table != derived.TableProjects || d.Id != "7" {
		t.Fatalf("projects/7, got %+v", d)
	}
	if app.calls != 2 {
		t.Fatalf("CreateTask 调用次数 = %d, want 2", app.calls)
	}
}
