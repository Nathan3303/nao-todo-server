// Package derived 收集一次请求内「服务端派生写」触及的行版本。
//
// 派生写指服务端因**计数联动 / 级联**等副作用推进了**非主写行**的 updated_at
// （如加检查项 ⇒ 父任务 check_item_count 与 updated_at 前进）。这类行没有独立回执，
// 客户端 base 因此静默过期 ⇒ 下次推父实体被 OCC 误判 stale（ADR 2026-09-28 RC-1）。
//
// 数据流：同步控制器在一次 /sync/push 注入**请求级** Recorder；dbs.TxManager.Do 在每个
// 顶层事务内建立**事务级** Recorder，仓储推进派生行时调用 Record；事务提交成功才并入
// 请求级（回滚则整批丢弃，避免把回滚写误报为可收敛版本）。控制器结束时去重回执。
//
// 非同步路径（REST）无请求级 Recorder ⇒ Record 为无副作用 no-op。
package derived

import (
	"context"
	"strconv"
	"time"
)

// 同步表名（与 /sync/push、/sync/pull 表键一致）。
const (
	TableTasks    = "tasks"
	TableProjects = "projects"
)

// Update 一次派生写触及的行与推进后的库中版本（毫秒精度，等于 DB datetime(3) 落库值）。
type Update struct {
	Table     string
	Id        string
	UpdatedAt time.Time
}

// Recorder 收集派生写；非并发安全（调用方按事务/请求串行使用）。
type Recorder struct {
	updates []Update
}

// NewRecorder 创建空收集器。
func NewRecorder() *Recorder { return &Recorder{} }

// Record 记录一次派生写（id <= 0 忽略）。
func (r *Recorder) Record(table string, id int64, at time.Time) {
	if r == nil || id <= 0 {
		return
	}
	r.updates = append(r.updates, Update{
		Table:     table,
		Id:        strconv.FormatInt(id, 10),
		UpdatedAt: at,
	})
}

// Updates 返回已收集派生写的副本（保持记录顺序）。
func (r *Recorder) Updates() []Update {
	if r == nil || len(r.updates) == 0 {
		return nil
	}
	out := make([]Update, len(r.updates))
	copy(out, r.updates)
	return out
}

// Dedupe 去重：同一 (table,id) 只保留**最后一次**记录的版本（即库中最终值），
// 并保持首次出现的位置顺序（回执稳定、可断言）。
func Dedupe(updates []Update) []Update {
	if len(updates) == 0 {
		return nil
	}
	index := make(map[Update]int, len(updates))
	out := make([]Update, 0, len(updates))
	for _, u := range updates {
		key := Update{Table: u.Table, Id: u.Id}
		if i, ok := index[key]; ok {
			out[i] = u // 覆盖为最终版本，位置不变
			continue
		}
		index[key] = len(out)
		out = append(out, u)
	}
	return out
}

type recorderKey struct{}

// WithRecorder 将收集器注入上下文。
func WithRecorder(ctx context.Context, r *Recorder) context.Context {
	return context.WithValue(ctx, recorderKey{}, r)
}

// Record 记录派生写；上下文无收集器时静默跳过（REST / 单测路径零影响）。
func Record(ctx context.Context, table string, id int64, at time.Time) {
	r, ok := ctx.Value(recorderKey{}).(*Recorder)
	if !ok {
		return
	}
	r.Record(table, id, at)
}

// Merge 把一次事务收集到的派生写并入上层（请求级）收集器；上层无收集器时静默跳过。
// 由 TxManager 在事务提交成功后调用（回滚不并入）。
func Merge(ctx context.Context, tx *Recorder) {
	if tx == nil {
		return
	}
	parent, ok := ctx.Value(recorderKey{}).(*Recorder)
	if !ok || parent == tx {
		return
	}
	parent.updates = append(parent.updates, tx.updates...)
}
