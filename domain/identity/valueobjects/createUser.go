package valueobjects

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"naotodoserver/domain/textutils"
)

type CreateUser struct {
	Account           string
	Password          string
	EncryptedPassword string
	Email             string
	Nickname          string
}

func (vo *CreateUser) Validate() error {
	if vo.Account == "" {
		return errors.New("账号不能为空")
	}
	if vo.Password == "" {
		return errors.New("密码不能为空")
	}
	if vo.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if vo.Nickname != "" && (textutils.RuneLength(vo.Nickname) < 2 || textutils.RuneLength(vo.Nickname) > 20) {
		return errors.New("昵称长度必须在 2-20 个字符之间")
	}
	return nil
}

func (vo *CreateUser) EncryptPassword() error {
	if vo.Password == "" {
		return errors.New("密码不能为空")
	}
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(vo.Password), 12)
	if err != nil {
		return errors.New("密码加密失败")
	}
	vo.EncryptedPassword = string(encryptedPassword)
	return nil
}

func NewCreateUser(account, password, email, nickname string) (*CreateUser, error) {
	vo := &CreateUser{
		Account:  account,
		Password: password,
		Email:    email,
		Nickname: nickname,
	}
	if err := vo.Validate(); err != nil {
		return nil, err
	}
	if err := vo.EncryptPassword(); err != nil {
		return nil, err
	}
	return vo, nil
}
