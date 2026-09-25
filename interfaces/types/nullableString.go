package types

import "encoding/json"

// NullableString 同步三态字符串：区分「absent（JSON 中无该键）」「null（显式置空）」与「值」。
//
// 存在原因（T318 根因 / T319 修复）：`*string` 无法区分 JSON null 与字段缺省（两者都绑成
// nil），客户端用 `null` 表达「清空可空时间」时被当成「不改」静默丢弃：覆盖分支跳过该列，
// 但同一请求仍写其它字段并 bump updated_at，回执仍是 applied ⇒ 客户端出队 ⇒ 下次 pull 用
// 服务端旧值覆盖本地（清单取消归档后任务又变回归档态）。
type NullableString struct {
	Present bool   // JSON 中出现该键（含 null）
	Null    bool   // JSON 值为 null（= 显式清空）
	Value   string // 非 null 时的字面值（可能为 ""）
}

// UnmarshalJSON 三态绑定：任何出现（含 null）都置 Present。
// @param data 原始 JSON 片段
// @return error 非 null 且非字符串时返回解码错误
func (n *NullableString) UnmarshalJSON(data []byte) error {
	n.Present = true
	if string(data) == "null" {
		n.Null = true
		n.Value = ""
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	n.Value = s
	return nil
}

// MarshalJSON 与 *string 同形回显：absent / null ⇒ null，其余 ⇒ 字符串
// （供载荷回显与契约测试断言，不参与业务写入）。
// @return []byte JSON 片段
// @return error 编码错误
func (n NullableString) MarshalJSON() ([]byte, error) {
	if !n.Present || n.Null {
		return []byte("null"), nil
	}
	return json.Marshal(n.Value)
}
