package controllers

// /sync/push「字段被静默丢弃」护栏（T466 / DEF-44）。
//
// 背景：服务端只承载 sync 条目 DTO 声明的字段，载荷中其余键被 encoding/json 静默丢弃，
// 却仍回执 applied ⇒ 客户端出队并误以为已同步（本次 P1 静默半写的放大器）。
//
// 口径（additive，⛔ 不改既有 outcome 语义）：
//   - 对每条推送记录，比对「原始 JSON 键集」与「sync 条目 DTO 可承载键集（含嵌入结构体）」
//     的差集，按字典序写入 SyncResult.DroppedFields；
//   - 键集由 interfaces/types 的 sync 条目结构体**反射 json tag** 得出，随 DTO 自动收敛
//     （字段一旦被承载即自动不再告警），不维护第二份手写清单；
//   - 服务端**只如实上报**，不拒绝、不改写 outcome；旧客户端忽略新增字段即无感。
//     ⚠️ 要真正阻止「出队即已同步」，需客户端消费该字段（跨仓改动）。

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"

	"naotodoserver/interfaces/types"
)

// syncTableItemTypes /sync/push 表键 → 其 sync 条目结构体类型（承载键集由此反射得出）。
var syncTableItemTypes = map[string]reflect.Type{
	"tasks":           reflect.TypeOf(types.SyncTaskPushItem{}),
	"taskCheckItems":  reflect.TypeOf(types.SyncCheckItemPushItem{}),
	"taskComments":    reflect.TypeOf(types.SyncCommentPushItem{}),
	"projects":        reflect.TypeOf(types.SyncProjectPushItem{}),
	"tags":            reflect.TypeOf(types.SyncTagPushItem{}),
	"pomodoros":       reflect.TypeOf(types.SyncPomodoroPushItem{}),
	"pomodoroRecords": reflect.TypeOf(types.SyncPomodoroRecordPushItem{}),
}

// syncTableCarriedKeys 预计算各表可承载的 JSON 键集（进程内只反射一次）。
var syncTableCarriedKeys = func() map[string]map[string]struct{} {
	m := make(map[string]map[string]struct{}, len(syncTableItemTypes))
	for table, typ := range syncTableItemTypes {
		m[table] = carriedJSONKeys(typ)
	}
	return m
}()

// carriedJSONKeys 收集结构体（递归展开匿名嵌入字段）在 JSON 层可承载的键集合，
// 忽略 json:"-" 与无 tag 字段。语义与 encoding/json 的字段提升一致。
//
// @param t 结构体类型（允许指针）
// @return 可承载的 JSON 键集合
func carriedJSONKeys(t reflect.Type) map[string]struct{} {
	set := make(map[string]struct{})
	var walk func(reflect.Type)
	walk = func(t reflect.Type) {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return
		}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.Anonymous {
				walk(f.Type)
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "" || name == "-" {
				continue
			}
			set[name] = struct{}{}
		}
	}
	walk(t)
	return set
}

// syncDroppedFields 承载一次 /sync/push 请求中逐表逐条被丢弃的 JSON 键。
type syncDroppedFields struct {
	byTable map[string][][]string
}

// newSyncDroppedFields 解析原始请求体，计算各表各条记录的丢弃键（字典序）。
// 解析失败或无表命中时返回空集对象（护栏静默降级，不影响主流程）。
//
// @param rawBody /sync/push 原始 JSON 请求体
// @return 丢弃键索引（非 nil）
func newSyncDroppedFields(rawBody []byte) *syncDroppedFields {
	d := &syncDroppedFields{byTable: make(map[string][][]string)}
	var raw map[string][]map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &raw); err != nil {
		return d
	}
	for table, records := range raw {
		carried, ok := syncTableCarriedKeys[table]
		if !ok {
			continue // 非七张 create 表（如 deletions）
		}
		perRecord := make([][]string, len(records))
		for i, record := range records {
			var extra []string
			for key := range record {
				if _, ok := carried[key]; !ok {
					extra = append(extra, key)
				}
			}
			if len(extra) > 0 {
				sort.Strings(extra)
				perRecord[i] = extra
			}
		}
		d.byTable[table] = perRecord
	}
	return d
}

// at 返回某表第 idx 条记录的丢弃键（越界或无丢弃返回 nil）。
//
// @param table /sync/push 表键
// @param idx 该表内记录下标（与绑定后的切片下标一致）
// @return 字典序丢弃键；无则 nil
func (d *syncDroppedFields) at(table string, idx int) []string {
	if d == nil {
		return nil
	}
	perRecord := d.byTable[table]
	if idx < 0 || idx >= len(perRecord) {
		return nil
	}
	return perRecord[idx]
}

// attach 按表内出现顺序（与推送循环逐条 append 的顺序一致）把丢弃键挂到对应回执条目上。
// ⛔ 只应传入七表 create 的结果切片；删除墓碑不参与。
//
// @param results 七表 create 的结果切片（原地写入）
func (d *syncDroppedFields) attach(results []types.SyncResult) {
	if d == nil {
		return
	}
	counters := make(map[string]int, len(syncTableItemTypes))
	for i := range results {
		table := results[i].Table
		idx := counters[table]
		counters[table]++
		if fields := d.at(table, idx); len(fields) > 0 {
			results[i].DroppedFields = fields
		}
	}
}
