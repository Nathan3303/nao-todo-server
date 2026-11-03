package types

import "time"

// DeleteDecision 删除（软删墓碑）的裁决结果（T642 / A4：删除纳入 OCC）
type DeleteDecision int

const (
	// DeleteApply 允许软删：写 deleted_at + updated_at（base 缺省 = 现行「删除优先」；base 命中 = OCC 通过）
	DeleteApply DeleteDecision = iota
	// DeleteStale OCC base 与库中 updated_at 不匹配 ⇒ 不写库、不产生新墓碑（R-F）
	DeleteStale
	// DeleteIdempotent 库中已是墓碑（同向重复删除）⇒ 幂等 applied，不写库、不产生新墓碑
	DeleteIdempotent
)

// DecideDelete 判定"记录已存在"时的删除动作（纯逻辑，便于单测）。
// 参数为库中当前 updated_at、库中是否已是墓碑、请求携带的 OCC base（零值 = 未提供）。
// 语义（与 DecideUpsert 同族）：
//   - base 缺省（零值）→ 回退现行「删除优先」（老客户端行为逐字不变）
//   - base 非零且与库中 updated_at 瞬时相等 → 允许软删
//   - base 非零且不相等 → Stale（不写入、不产生新墓碑，由调用方回传库中当前版本供客户端 rebase）
//   - 库中已是墓碑 → 幂等（不写入、不推进 updated_at）
//
// @param existingUpdated 库中当前 updated_at
// @param existingDeleted 库中是否已是墓碑
// @param baseUpdated 客户端回传的 OCC base（零值 = 未提供）
// @return DeleteDecision 删除动作
func DecideDelete(
	existingUpdated time.Time,
	existingDeleted bool,
	baseUpdated time.Time,
) DeleteDecision {
	if !baseUpdated.IsZero() && !baseUpdated.Equal(existingUpdated) {
		return DeleteStale
	}
	if existingDeleted {
		return DeleteIdempotent
	}
	return DeleteApply
}
