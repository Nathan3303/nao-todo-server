package entities

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"naotodoserver/domain/textutils"
	"naotodoserver/domain/types"
)

// User 用户实体
// 用于表示用户在系统中的身份和权限
// 包含用户账号、邮箱、密码、昵称、头像、创建来源、角色、状态、配置等属性
type User struct {
	types.EntityBase
	Account     string
	Email       string
	Password    string
	Nickname    string
	Avatar      string
	CreatedFrom string
	Role        UserRole
	State       UserState
	DeactivedAt types.NullableTime
	Config      *UserConfig
}

// IsIdValid 检查用户 ID 是否有效
func (u *User) IsIdValid() bool {
	return u != nil && u.Id > 0
}

// IsNicknameValid 检查用户昵称是否有效
func (u *User) IsNicknameValid() bool {
	return u != nil &&
		textutils.RuneLength(u.Nickname) >= 2 &&
		textutils.RuneLength(u.Nickname) <= 20
}

// IsValid 检查用户实体是否有效
func (u *User) IsValid() bool {
	return u.IsIdValid() && u.IsNicknameValid()
}

// EncryptPassword 加密用户密码
func (u *User) EncryptPassword() error {
	if u.Password == "" {
		return errors.New("密码不能为空")
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return errors.New("密码加密失败")
	}
	u.Password = string(encryptedPassword)
	return nil
}

// IsDeactived 检查用户是否已停用
// 如果 DeactivedAt 不为空，则用户已停用
func (u *User) IsDeactived() bool {
	return u.DeactivedAt.IsNull
}
