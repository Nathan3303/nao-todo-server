package valueobjects

import (
	"errors"

	"naotodoserver/domain/types"
)

// SaveProjectPreference 保存项目偏好值对象
type SaveProjectPreference struct {
	ViewType   types.ViewType
	GetOptions string
	Columns    string
}

// Validate 验证保存项目偏好值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (saveProjectPreference *SaveProjectPreference) Validate() error {
	if saveProjectPreference.ViewType == "" {
		return errors.New("视图类型不能为空")
	}
	if saveProjectPreference.ViewType != types.ViewTypeTable &&
		saveProjectPreference.ViewType != types.ViewTypeList &&
		saveProjectPreference.ViewType != types.ViewTypeKanban {
		return errors.New("视图类型必须是 table、list 或 kanban")
	}
	if saveProjectPreference.GetOptions == "" {
		return errors.New("获取任务选项不能为空")
	}
	if saveProjectPreference.Columns == "" {
		return errors.New("列选项不能为空")
	}
	return nil
}

// NewSaveProjectPreference 创建保存项目偏好值对象
// @param viewType 视图类型
// @param getOptions 获取任务选项
// @param columns 列选项
// @return *SaveProjectPreference 保存项目偏好值对象
// @return error 验证失败返回错误，否则返回 nil
func NewSaveProjectPreference(
	viewType types.ViewType,
	getOptions string,
	columns string,
) (*SaveProjectPreference, error) {
	vo := &SaveProjectPreference{
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
