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
)

// DecideUpsert 判定"记录已存在"时的 upsert 动作（纯逻辑，便于单测）。
// 参数为库中记录的 created_at/updated_at 与请求携带的 createdAt/updatedAt（零值表示未提供）。
// conflictWindow 为 create 语义冲突判定窗口：请求携带 createdAt 且与库中相差超过该窗口 → 判为 ID 碰撞（返回 ErrIDConflict）。
// LWW 判定：请求 updatedAt 非零且严格早于库中 updated_at → Noop；否则 Overwrite（updatedAt 零值也视为覆盖，服务端取 now）。
func DecideUpsert(
	existingCreated, existingUpdated, voCreated, voUpdated time.Time,
	conflictWindow time.Duration,
) (UpsertOutcome, error) {
	if !voCreated.IsZero() && existingCreated.Sub(voCreated).Abs() > conflictWindow {
		return UpsertConflict, domerr.ErrIDConflict
	}
	if !voUpdated.IsZero() && existingUpdated.After(voUpdated) {
		return UpsertNoop, nil
	}
	return UpsertOverwrite, nil
}
