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
		outcome, err := DecideUpsert(base, base, base.Add(-2*time.Minute), time.Time{}, win)
		if outcome != UpsertConflict || !errors.Is(err, domerr.ErrIDConflict) {
			t.Fatalf("outcome=%v err=%v, want Conflict/ErrIDConflict", outcome, err)
		}
	})

	t.Run("create 语义重试：createdAt 与库中接近 → 不冲突，走 LWW", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, base.Add(-10*time.Second), base, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("请求 updatedAt 更旧 → Noop（幂等重试不覆盖新数据）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base.Add(5*time.Minute), time.Time{}, base, win)
		if err != nil || outcome != UpsertNoop {
			t.Fatalf("outcome=%v err=%v, want Noop", outcome, err)
		}
	})

	t.Run("请求 updatedAt 相同 → Overwrite（相等时远程胜）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("请求 updatedAt 更新 → Overwrite", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base.Add(1*time.Minute), win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("未提供 updatedAt → Overwrite（服务端取 now）", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base.Add(1*time.Hour), time.Time{}, time.Time{}, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})

	t.Run("未提供 createdAt（纯 update 语义）→ 不触发冲突检测", func(t *testing.T) {
		outcome, err := DecideUpsert(base, base, time.Time{}, base, win)
		if err != nil || outcome != UpsertOverwrite {
			t.Fatalf("outcome=%v err=%v, want Overwrite", outcome, err)
		}
	})
}
