package types

import "reflect"

// CopyStruct copies fields from src to dst for structs with identical field
// names and compatible types.
// Only exported fields are copied. Fields in dst that are zero-valued in src are left unchanged.
// This is a shallow copy — reference types (slices, maps, pointers) share the underlying data.
func CopyStruct(src, dst any) {
	srcVal := reflect.ValueOf(src).Elem()
	dstVal := reflect.ValueOf(dst).Elem()

	for i := range srcVal.NumField() {
		srcField := srcVal.Field(i)
		srcType := srcVal.Type().Field(i)
		if !srcType.IsExported() {
			continue
		}
		dstField := dstVal.FieldByName(srcType.Name)
		if !dstField.IsValid() || dstField.Kind() != srcField.Kind() {
			continue
		}
		if dstField.CanSet() {
			dstField.Set(srcField)
		}
	}
}
