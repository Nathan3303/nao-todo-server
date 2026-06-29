package types

import (
	"bytes"
	"encoding/json"
)

// NullableString 可空字符串
type NullableString struct {
	Present bool
	IsNull  bool
	Value   string
}

// UnmarshalJSON 反序列化 JSON 数据
func (ns *NullableString) UnmarshalJSON(data []byte) error {
	ns.Present = true
	if bytes.Equal(data, []byte("null")) {
		ns.IsNull = true
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	ns.Value = s
	return nil
}

// ToUpdateState 将 NullableString 转换为更新状态
// @return shouldUpdate 是否需要更新
// @return isNull 是否设置为 NULL
// @return value 值
func (ns NullableString) ToUpdateState() (shouldUpdate bool, isNull bool, value string) {
	if !ns.Present {
		return false, false, ""
	}
	if ns.IsNull {
		return true, true, ""
	}
	return true, false, ns.Value
}
