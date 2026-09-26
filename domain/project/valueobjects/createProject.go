package valueobjects

import (
	"errors"
	"time"

	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// CreateProject 创建项目值对象
type CreateProject struct {
	Id          int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   types.NullableTime
	UserId      int64
	Name        string
	Description string
	SortId      uint16

	// ArchivedAt 归档时间三态（sync push 专用，T322 / DEF-42）：
	// Valid=false（缺省；REST create 与旧客户端推送）⇒ 不写列；
	// Valid=true,IsNull=true（null / ""）⇒ 显式清空写 NULL；Valid=true,IsNull=false ⇒ 写入该时间。
	// ⛔ 服务端只做行级落地，不级联归档清单下的任务（级联由客户端逐任务推送完成）。
	ArchivedAt types.NullableTime

	// BaseUpdatedAt OCC：客户端回传的服务端 updated_at 快照（零值 = 未提供，回退 LWW）
	BaseUpdatedAt time.Time
}

// Validate 验证创建项目值对象是否符合要求
// @return error 验证失败返回错误，否则返回 nil
func (createVO *CreateProject) Validate() error {
	// 验证用户 ID 是否为空
	if createVO.UserId == 0 {
		return errors.New("用户 ID 不能为空")
	}
	// 验证项目名称是否为空
	if createVO.Name == "" {
		return errors.New("项目名称不能为空")
	}
	// 验证项目名称长度是否超过128个字符
	if textutils.RuneLength(createVO.Name) > 128 {
		return errors.New("项目名称不能超过128个字符")
	}
	// 验证项目描述是否超过256个字符
	if createVO.Description != "" && textutils.RuneLength(createVO.Description) > 512 {
		return errors.New("项目描述不能超过512个字符")
	}
	return nil
}

// NewCreateProject 创建项目值对象
// @param userId 用户 ID
// @param name 项目名称
// @param description 项目描述
// @return *CreateProject 创建项目值对象
// @return error 验证失败返回错误，否则返回 nil
func NewCreateProject(userId int64, name string, description string) (*CreateProject, error) {
	vo := &CreateProject{
		UserId:      userId,
		Name:        name,
		Description: description,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
