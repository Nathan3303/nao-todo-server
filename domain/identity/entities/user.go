package entities

import (
	"errors"
	"time"

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
	LastCancelRestoreAt types.NullableTime
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
	return !u.DeactivedAt.IsNull
}

// IsInCooldown 检查用户是否处于注销冷却期
// 如果 LastCancelRestoreAt 不为空且距今不足30天，则处于冷却期
func (u *User) IsInCooldown() bool {
	t, ok := u.LastCancelRestoreAt.Value()
	if !ok {
		return false
	}
	return time.Since(t) < 30*24*time.Hour
}

// CooldownRemainingDays 获取冷却期剩余天数
// 返回剩余整天数，若不在冷却期则返回0
func (u *User) CooldownRemainingDays() int {
	t, ok := u.LastCancelRestoreAt.Value()
	if !ok {
		return 0
	}
	remaining := 30*24*time.Hour - time.Since(t)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Hours()/24) + 1
}
