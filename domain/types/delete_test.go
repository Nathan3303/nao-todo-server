package types

import (
	"testing"
	"time"
)

// TestDecideDelete 删除裁决四分支（T642 / A4）：
// 无 base（老客户端）/ base 命中 / base 不匹配 / 已是墓碑（同向重复删除）。
func TestDecideDelete(t *testing.T) {
	dbUpdated := time.Date(2026, 11, 3, 10, 0, 0, 123e6, time.UTC)

	t.Run("无 base（老客户端）⇒ 允许软删，行为与今天一致", func(t *testing.T) {
		if got := DecideDelete(dbUpdated, false, time.Time{}); got != DeleteApply {
			t.Fatalf("got=%v, want DeleteApply", got)
		}
	})

	t.Run("base == 库中 updated_at ⇒ 允许软删", func(t *testing.T) {
		if got := DecideDelete(dbUpdated, false, dbUpdated); got != DeleteApply {
			t.Fatalf("got=%v, want DeleteApply", got)
		}
	})

	t.Run("base != 库中 updated_at ⇒ Stale（不写库、不产生新墓碑）", func(t *testing.T) {
		for _, base := range []time.Time{
			dbUpdated.Add(-time.Millisecond),
			dbUpdated.Add(time.Millisecond),
			dbUpdated.Add(-time.Hour),
		} {
			if got := DecideDelete(dbUpdated, false, base); got != DeleteStale {
				t.Fatalf("base=%v got=%v, want DeleteStale", base, got)
			}
		}
	})

	t.Run("已是墓碑（同向重复删除）⇒ 幂等，不写库", func(t *testing.T) {
		if got := DecideDelete(dbUpdated, true, time.Time{}); got != DeleteIdempotent {
			t.Fatalf("无 base got=%v, want DeleteIdempotent", got)
		}
		if got := DecideDelete(dbUpdated, true, dbUpdated); got != DeleteIdempotent {
			t.Fatalf("base 命中 got=%v, want DeleteIdempotent", got)
		}
	})

	t.Run("已是墓碑但 base 不匹配 ⇒ Stale（版本已变，交客户端裁决）", func(t *testing.T) {
		if got := DecideDelete(dbUpdated, true, dbUpdated.Add(-time.Second)); got != DeleteStale {
			t.Fatalf("got=%v, want DeleteStale", got)
		}
	})
}
