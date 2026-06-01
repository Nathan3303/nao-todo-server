package valueobjects

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"naotodoserver/domain/textutils"
)

// CreateUser 创建用户值对象
type CreateUser struct {
	Account           string
	Password          string
	EncryptedPassword string
	Email             string
	Nickname          string
}

// Validate 验证创建用户值对象
// @return error 错误信息
func (createUser *CreateUser) Validate() error {
	if createUser.Account == "" {
		return errors.New("账号不能为空")
	}
	if createUser.Password == "" {
		return errors.New("密码不能为空")
	}
	if createUser.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if createUser.Nickname != "" && (textutils.RuneLength(createUser.Nickname) < 2 || textutils.RuneLength(createUser.Nickname) > 20) {
		return errors.New("昵称长度必须在 2-20 个字符之间")
	}
	return nil
}

// EncryptPassword 加密密码
// @return error 错误信息
func (createUser *CreateUser) EncryptPassword() error {
	// 密码不能为空
	if createUser.Password == "" {
		return errors.New("密码不能为空")
	}
	// 密码加密
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(createUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("密码加密失败")
	}
	// 密码赋值
	createUser.EncryptedPassword = string(encryptedPassword)
	return nil

}

// NewCreateUser 创建创建用户值对象
// @description 创建创建用户值对象，验证并加密密码
// @return *CreateUser 创建用户值对象
// @return error 错误信息
func NewCreateUser(account, password, email, nickname string) (*CreateUser, error) {
	vo := &CreateUser{
		Account:  account,
		Password: password,
		Email:    email,
		Nickname: nickname,
	}
	validateErr := vo.Validate()
	if validateErr != nil {
		return nil, validateErr
	}
	encryptErr := vo.EncryptPassword()
	if encryptErr != nil {
		return nil, encryptErr
	}
	return vo, nil
}
