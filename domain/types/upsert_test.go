package types

import (
	"errors"
	"testing"
	"time"

	domerr "naotodoserver/domain/errors"
)

func TestDecideUpsert(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win := time.Minute

	t.Run("create 语义碰撞：createdAt 与库中相差 > 1 分钟 → Conflict", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, base.Add(-2*time.Minute), time.Time{}, time.Time{}, win)
		if outcome != UpsertConflict || !errors.Is(err, domerr.ErrIDConflict) {
			t.Fatalf("outcome=%v err=%v, want Conflict/ErrIDConflict", outcome, err)
		}
	})

	t.Run("create 语义重试：createdAt 与库中接近 → 不冲突，走 LWW", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, base.Add(-10*time.Second), base, time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("请求 updatedAt 更旧 → Noop（幂等重试不覆盖新数据）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base.Add(5*time.Minute), time.Time{}, base, time.Time{}, win)
		if err != nil || outcome != UpsertNoop {
			t.Fatalf("outcome=%v err=%v, want Noop", outcome, err)
		}
	})

	t.Run("请求 updatedAt 相同 → Overwrite（相等时远程胜）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base, time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("请求 updatedAt 更新 → Overwrite", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base.Add(1*time.Minute), time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("未提供 updatedAt → Overwrite（服务端取 now）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base.Add(1*time.Hour), time.Time{}, time.Time{}, time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("未提供 createdAt（纯 update 语义）→ 不触发冲突检测", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base, time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})
}

// TestDecideUpsertOCCBase T163 契约：baseUpdatedAt（服务端 updated_at 快照）分支。
// base 缺失 ⇒ 现行 LWW 逐字不变；base 相等 ⇒ 覆盖；base 不等 ⇒ stale（不写、回库中版本）。
func TestDecideUpsertOCCBase(t *testing.T) {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	win := time.Minute

	t.Run("base 缺失（零值）且请求更旧 → 仍为 Noop（LWW 零回归）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base.Add(5*time.Minute), time.Time{}, base, time.Time{}, win)
		if err != nil || outcome != UpsertNoop {
			t.Fatalf("outcome=%v err=%v, want Noop", outcome, err)
		}
	})

	t.Run("base 缺失（零值）且请求更新 → 仍为 Overwrite（LWW 零回归）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base.Add(time.Minute), time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("base 与库中版本瞬时相等（不同时区表示）→ Overwrite", func(t *testing.T) {
		baseCST := base.In(time.FixedZone("CST", 8*3600))
		outcome, err := DecideUpsert(base, base, time.Time{}, base.Add(-time.Hour), baseCST, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite（Equal 忽略时区）", outcome, err)
		}
	})

	t.Run("base 与库中版本不等 → Stale（即使请求 updatedAt 更新）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base.Add(10*time.Minute), base.Add(-time.Minute), win)
		if err != nil || outcome != UpsertStale {
			t.Fatalf("outcome=%v err=%v, want Stale", outcome, err)
		}
	})

	t.Run("base 不等但 createdAt 相差过大 → Conflict 优先（ID 碰撞语义不变）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, base.Add(-2*time.Minute), base, base.Add(-time.Minute), win)
		if outcome != UpsertConflict || !errors.Is(err, domerr.ErrIDConflict) {
			t.Fatalf("outcome=%v err=%v, want Conflict/ErrIDConflict", outcome, err)
		}
	})
}
