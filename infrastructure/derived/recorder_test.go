package derived

import (
	"context"
	"testing"
	"time"
)

// TestRecord_NoRecorderIsNoop 非同步路径（无请求级收集器）下 Record 必须零副作用。
func TestRecord_NoRecorderIsNoop(t *testing.T) {
	Record(context.Background(), TableTasks, 42, time.Now()) // 不应 panic
}

// TestRecordAndUpdates 记录顺序保持；id<=0 忽略。
func TestRecordAndUpdates(t *testing.T) {
	r := NewRecorder()
	at1 := time.Date(2026, 9, 28, 10, 0, 0, 123_000_000, time.UTC)
	at2 := at1.Add(5 * time.Millisecond)
	r.Record(TableTasks, 42, at1)
	r.Record(TableTasks, 0, at2) // 忽略
	r.Record(TableProjects, 7, at2)

	got := r.Updates()
	if len(got) != 2 {
		t.Fatalf("Updates 条数 = %d, want 2: %+v", len(got), got)
	}
	if got[0].Table != TableTasks || got[0].Id != "42" || !got[0].UpdatedAt.Equal(at1) {
		t.Fatalf("首条 = %+v, want tasks/42/%v", got[0], at1)
	}
	if got[1].Table != TableProjects || got[1].Id != "7" {
		t.Fatalf("次条 = %+v, want projects/7", got[1])
	}
	// 返回副本：外部修改不回流
	got[0].Id = "changed"
	if r.Updates()[0].Id != "42" {
		t.Fatal("Updates 应返回副本")
	}
}

// TestDedupe_KeepsLastValueAndFirstPosition 同 (table,id) 去重保留最终版本，位置保持首次出现顺序。
func TestDedupe_KeepsLastValueAndFirstPosition(t *testing.T) {
	t1 := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Millisecond)
	t3 := t2.Add(time.Millisecond)
	in := []Update{
		{Table: TableTasks, Id: "1", UpdatedAt: t1},
		{Table: TableProjects, Id: "9", UpdatedAt: t1},
		{Table: TableTasks, Id: "1", UpdatedAt: t2},
		{Table: TableTasks, Id: "2", UpdatedAt: t2},
		{Table: TableTasks, Id: "1", UpdatedAt: t3},
	}
	got := Dedupe(in)
	if len(got) != 3 {
		t.Fatalf("Dedupe 条数 = %d, want 3: %+v", len(got), got)
	}
	if got[0].Table != TableTasks || got[0].Id != "1" || !got[0].UpdatedAt.Equal(t3) {
		t.Fatalf("首条应为 tasks/1 的最终版本 %v, got %+v", t3, got[0])
	}
	if got[1].Table != TableProjects || got[1].Id != "9" {
		t.Fatalf("次条位置应保持首次出现顺序 = projects/9, got %+v", got[1])
	}
	if got[2].Id != "2" {
		t.Fatalf("末条应为 tasks/2, got %+v", got[2])
	}
	if Dedupe(nil) != nil {
		t.Fatal("空输入应返回 nil")
	}
}

// TestMerge_CommitsIntoParentAndSelfGuard 事务级收集器在提交后并入请求级；自并入无副作用。
func TestMerge_CommitsIntoParentAndSelfGuard(t *testing.T) {
	parent := NewRecorder()
	txCtx := WithRecorder(context.Background(), parent)
	tx := NewRecorder()
	at := time.Date(2026, 9, 28, 10, 0, 0, 0, time.UTC)
	tx.Record(TableTasks, 5, at)

	Merge(txCtx, tx)
	if got := parent.Updates(); len(got) != 1 || got[0].Id != "5" {
		t.Fatalf("Merge 应把事务收集器并入请求级收集器, got %+v", got)
	}
	// 无请求级收集器 ⇒ 静默跳过
	Merge(context.Background(), tx)
	// 自并入守卫
	tx2 := NewRecorder()
	Merge(WithRecorder(context.Background(), tx2), tx2)
	if len(tx2.Updates()) != 0 {
		t.Fatalf("自并入应被守卫, got %+v", tx2.Updates())
	}
}
