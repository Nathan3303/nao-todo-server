package types

import (
	"time"

	domerr "naotodoserver/domain/errors"
)

// UpsertOutcome upsert 判定结果
type UpsertOutcome int

const (
	// UpsertOverwrite 请求不旧于库中版本 → 覆盖写入
	UpsertOverwrite UpsertOutcome = iota
	// UpsertNoop 请求更旧 → 不写入，返回库中当前版本
	UpsertNoop
	// UpsertConflict create 语义且 created_at 相差过大 → 判定为不同实体 ID 碰撞
	UpsertConflict
	// UpsertStale OCC base 版本与库中当前版本不匹配 → 不写入，返回库中当前版本（T163）
	UpsertStale
)

// UpsertResult 单次幂等写入（upsert）的实际动作。
// Outcome 与 DecideUpsert 判定同源（ID 冲突以 error 返回，不落入本结构）；
// Created 表示本次为新建（含墓碑复活，B6），供计数事件 / 附属记录初始化判定使用。
type UpsertResult struct {
	Outcome UpsertOutcome
	Created bool
}

// DecideUpsert 判定"记录已存在"时的 upsert 动作（纯逻辑，便于单测）。
// 参数为库中记录的 created_at/updated_at 与请求携带的 createdAt/updatedAt（零值表示未提供）。
// conflictWindow 为 create 语义冲突判定窗口：请求携带 createdAt 且与库中相差超过该窗口 → 判为 ID 碰撞（返回 ErrIDConflict）。
// baseUpdated 为客户端回传的 OCC 版本快照（服务端 updated_at，零值表示未提供）：
//   - 零值 → 维持现行 LWW（请求 updatedAt 非零且严格早于库中 updated_at → Noop；否则 Overwrite）
//   - 非零且与库中 updated_at 瞬时相等 → Overwrite
//   - 非零且不相等 → Stale（不写入，由仓储回传库中当前版本供客户端 rebase）
func DecideUpsert(
	existingCreated, existingUpdated, voCreated, voUpdated, baseUpdated time.Time,
	conflictWindow time.Duration,
) (UpsertOutcome, error) {
	if !voCreated.IsZero() && existingCreated.Sub(voCreated).Abs() > conflictWindow {
		return UpsertConflict, domerr.ErrIDConflict
	}
	// OCC 分支：base 缺失（零值）即回退现行 LWW，保证 2A 客户端与其它调用方零行为变化
	if !baseUpdated.IsZero() {
		if baseUpdated.Equal(existingUpdated) {
			return UpsertOverwrite, nil
		}
		return UpsertStale, nil
	}
	if !voUpdated.IsZero() && existingUpdated.After(voUpdated) {
		return UpsertNoop, nil
	}
	return UpsertOverwrite, nil
}
