package valueobjects

import (
	"errors"
)

// SaveTagPreference 保存标签偏好值对象
type SaveTagPreference struct {
	ViewType   string
	GetOptions string
	Columns    string
}

// Validate 验证保存标签偏好值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (saveTagPreference *SaveTagPreference) Validate() error {
	if saveTagPreference.ViewType == "" {
		return errors.New("视图类型不能为空")
	}
	if saveTagPreference.ViewType != "table" &&
		saveTagPreference.ViewType != "list" &&
		saveTagPreference.ViewType != "kanban" {
		return errors.New("视图类型必须是 table、list 或 kanban")
	}
	if saveTagPreference.GetOptions == "" {
		return errors.New("获取任务选项不能为空")
	}
	if saveTagPreference.Columns == "" {
		return errors.New("列选项不能为空")
	}
	return nil
}

// NewSaveTagPreference 创建保存标签偏好值对象
// @param viewType 视图类型
// @param getOptions 获取任务选项
// @param columns 列选项
// @return *SaveTagPreference 保存标签偏好值对象
// @return error 验证失败返回错误，否则返回 nil
func NewSaveTagPreference(
	viewType string,
	getOptions string,
	columns string,
) (*SaveTagPreference, error) {
	vo := &SaveTagPreference{
		ViewType:   viewType,
		GetOptions: getOptions,
		Columns:    columns,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
