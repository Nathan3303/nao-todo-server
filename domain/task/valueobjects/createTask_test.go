package valueobjects

import (
	"strings"
	"testing"
	"time"

	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/types"
)

// newStartAtTask 走 NewCreateTask 真实构造路径（含 FillStartAt 兜底）
func newStartAtTask(t *testing.T, startAt, endAt *string) *CreateTask {
	t.Helper()
	vo, err := NewCreateTask(
		0, "任务", "", entities.TaskStatePending, entities.TaskPriorityMedium,
		startAt, endAt, 0, nil, nil, nil, nil, nil, 0, "", 0,
	)
	if err != nil {
		t.Fatalf("NewCreateTask: %v", err)
	}
	return vo
}

// --- DEF-SYNC-04：创建时不得覆盖客户端显式 startAt ---

// ① 提供 startAt+endAt（endAt 未来）⇒ startAt 原样保留，不被服务端 now 覆盖
func TestNewCreateTask_ProvidedStartAtNotOverwritten(t *testing.T) {
	startAt := time.Now().Add(time.Hour).Truncate(time.Second)
	endAt := startAt.Add(2 * time.Hour)
	vo := newStartAtTask(t, strPtr(startAt.Format(time.RFC3339)), strPtr(endAt.Format(time.RFC3339)))
	if vo.StartAt.IsNull {
		t.Fatal("startAt 应为有效值")
	}
	if !vo.StartAt.Time.Equal(startAt) {
		t.Fatalf("显式 startAt 被覆盖: got %v, want %v", vo.StartAt.Time, startAt)
	}
	if !vo.EndAt.Time.Equal(endAt) {
		t.Fatalf("endAt 被改写: got %v, want %v", vo.EndAt.Time, endAt)
	}
}

// ② endAt < now 且客户端提供 startAt ⇒ 保留客户端值，不再产出 startAt > endAt 倒置
func TestNewCreateTask_EndAtPastNoInversion(t *testing.T) {
	// 旧 FillStartAt：endAt(过去) ⇒ startAt = now−1min ⇒ 必然 > endAt(更早过去) ⇒ 倒置
	startAt := time.Now().Add(-3 * time.Minute).Truncate(time.Second)
	endAt := startAt.Add(time.Minute) // endAt 仍早于 now，但晚于 startAt
	vo := newStartAtTask(t, strPtr(startAt.Format(time.RFC3339)), strPtr(endAt.Format(time.RFC3339)))
	if vo.StartAt.IsNull {
		t.Fatal("startAt 应为有效值")
	}
	if !vo.StartAt.Time.Equal(startAt) {
		t.Fatalf("显式 startAt 被覆盖: got %v, want %v", vo.StartAt.Time, startAt)
	}
	if vo.StartAt.Time.After(vo.EndAt.Time) {
		t.Fatalf("产出 startAt(%v) > endAt(%v) 倒置", vo.StartAt.Time, vo.EndAt.Time)
	}
}

// ③a 缺省 startAt + endAt 有效（未来）⇒ 既有兜底语义不变（派生 startAt ≈ 创建时刻）
func TestNewCreateTask_MissingStartAtFutureEndAtFallback(t *testing.T) {
	endAt := time.Now().Add(time.Hour).Truncate(time.Second)
	before := time.Now()
	vo := newStartAtTask(t, nil, strPtr(endAt.Format(time.RFC3339)))
	after := time.Now()
	if vo.StartAt.IsNull {
		t.Fatal("缺省 startAt 应走兜底派生")
	}
	if vo.StartAt.Time.Before(before.Add(-time.Second)) || vo.StartAt.Time.After(after.Add(time.Second)) {
		t.Fatalf("派生 startAt 应≈创建时刻: got %v, 窗口 [%v, %v]", vo.StartAt.Time, before, after)
	}
}

// ③b 缺省 startAt + 缺省 endAt ⇒ 两者皆空（未安排），既有语义不变
func TestNewCreateTask_MissingStartAtMissingEndAt(t *testing.T) {
	vo := newStartAtTask(t, nil, nil)
	if !vo.StartAt.IsNull {
		t.Fatal("缺省 startAt 应为空")
	}
	if !vo.EndAt.IsNull {
		t.Fatal("缺省 endAt 应为空")
	}
}

// ③c 无效 startAt（不可解析）视同缺省 ⇒ 走兜底派生（endAt 有效）
func TestNewCreateTask_InvalidStartAtFallsBack(t *testing.T) {
	endAt := time.Now().Add(time.Hour).Truncate(time.Second)
	before := time.Now()
	vo := newStartAtTask(t, strPtr("not-a-date"), strPtr(endAt.Format(time.RFC3339)))
	after := time.Now()
	if vo.StartAt.IsNull {
		t.Fatal("无效 startAt 应视同缺省并兜底派生")
	}
	if vo.StartAt.Time.Before(before.Add(-time.Second)) || vo.StartAt.Time.After(after.Add(time.Second)) {
		t.Fatalf("派生 startAt 应≈创建时刻: got %v, 窗口 [%v, %v]", vo.StartAt.Time, before, after)
	}
}

// SYNC-DEF-01：显式空串 startAt（清空）+ endAt 有效 ⇒ 不得被 FillStartAt 复活为 now，
// 应保持「已提供且置空」语义，由持久化层写 NULL。
func TestNewCreateTask_ExplicitEmptyStartAtStaysCleared(t *testing.T) {
	endAt := time.Now().Add(time.Hour).Truncate(time.Second)
	vo := newStartAtTask(t, strPtr(""), strPtr(endAt.Format(time.RFC3339)))
	if !vo.StartAt.Valid || !vo.StartAt.IsNull {
		t.Fatalf("显式空串 startAt 应保持清空语义（Valid=true,IsNull=true），got Valid=%v IsNull=%v Time=%v",
			vo.StartAt.Valid, vo.StartAt.IsNull, vo.StartAt.Time)
	}
	if !vo.EndAt.Time.Equal(endAt) {
		t.Fatalf("endAt 被改写: got %v, want %v", vo.EndAt.Time, endAt)
	}
}

// SYNC-DEF-01：缺省与显式空串必须可区分（Valid 位）
func TestNewCreateTask_AbsentVsExplicitEmpty(t *testing.T) {
	absent := newStartAtTask(t, nil, nil)
	if absent.StartAt.Valid {
		t.Fatal("缺省 startAt 应为 Valid=false（未提供）")
	}
	cleared := newStartAtTask(t, strPtr(""), nil)
	if !cleared.StartAt.Valid || !cleared.StartAt.IsNull {
		t.Fatalf("显式空串 startAt 应为 Valid=true,IsNull=true，got Valid=%v IsNull=%v",
			cleared.StartAt.Valid, cleared.StartAt.IsNull)
	}
}

// FillStartAt 直接守卫：已提供非空 startAt ⇒ no-op（无论 endAt 取值）
func TestFillStartAt_ProvidedStartAtNoop(t *testing.T) {
	startAt := time.Now().Add(time.Hour).Truncate(time.Second)
	endAt := startAt.Add(time.Hour)
	ct := &CreateTask{
		StartAt: types.NewNullableTimeByTime(startAt),
		EndAt:   types.NewNullableTimeByTime(endAt),
	}
	ct.FillStartAt()
	if !ct.StartAt.Time.Equal(startAt) {
		t.Fatalf("FillStartAt 覆盖了显式 startAt: got %v, want %v", ct.StartAt.Time, startAt)
	}
}

func TestCreateTask_Validate(t *testing.T) {
	longName := strings.Repeat("a", 257)
	longDesc := strings.Repeat("b", 513)

	tests := []struct {
		name    string
		vo      *CreateTask
		wantErr string
	}{
		{
			name: "valid task",
			vo:   &CreateTask{Name: "Valid Task", Description: "A description"},
		},
		{
			name:    "empty name",
			vo:      &CreateTask{Name: "", Description: "desc"},
			wantErr: "任务名称不能为空",
		},
		{
			name:    "name too long",
			vo:      &CreateTask{Name: longName},
			wantErr: "任务名称最多256个字符",
		},
		{
			name:    "description too long",
			vo:      &CreateTask{Name: "Valid", Description: longDesc},
			wantErr: "任务描述最多512个字符",
		},
		{
			name: "optional description empty",
			vo:   &CreateTask{Name: "Valid", Description: ""},
		},
		{
			name: "max length name",
			vo:   &CreateTask{Name: strings.Repeat("a", 256)},
		},
		{
			name: "max length description",
			vo:   &CreateTask{Name: "Valid", Description: strings.Repeat("c", 512)},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.vo.Validate()
			if tt.wantErr == "" && err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
			if tt.wantErr != "" && err == nil {
				t.Errorf("Validate() want error %q, got nil", tt.wantErr)
			}
			if tt.wantErr != "" && err != nil && err.Error() != tt.wantErr {
				t.Errorf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
