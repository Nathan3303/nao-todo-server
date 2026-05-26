package types

import (
	"bytes"
	"encoding/json"
)

type NullableString struct {
	Present bool
	IsNull  bool
	Value   string
}

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

func (ns NullableString) ToUpdateState() (shouldUpdate bool, isNull bool, value string) {
	if !ns.Present {
		return false, false, ""
	}
	if ns.IsNull {
		return true, true, ""
	}
	return true, false, ns.Value
}
