package valueobjects

import "errors"

// 保存任务清单偏好值对象
type SaveProjectPreference struct {
	ViewType   string `json:"viewType"`
	GetOptions string `json:"getTodosOptions"`
	Columns    string `json:"columns"`
}

// 验证保存任务清单偏好值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (saveProjectPreference *SaveProjectPreference) Validate() error {
	if saveProjectPreference.ViewType == "" {
		return errors.New("视图类型不能为空")
	}
	if saveProjectPreference.ViewType != "table" &&
		saveProjectPreference.ViewType != "list" &&
		saveProjectPreference.ViewType != "kanban" {
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

// 创建保存任务清单偏好值对象
// @param viewType 视图类型
// @param getOptions 获取任务选项
// @param columns 列选项
// @return *SaveProjectPreference 保存任务清单偏好值对象
// @return error 验证失败返回错误，否则返回 nil
func NewSaveProjectPreference(
	viewType string,
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
