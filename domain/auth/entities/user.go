package entities

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id       int64  `json:"id"`
	Account  string `json:"account"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

func (u *User) IsIdValid() bool {
	return u != nil && u.Id > 0
}

func (u *User) IsNicknameValid() bool {
	return u != nil && len(u.Nickname) >= 2 && len(u.Nickname) <= 20
}

func (u *User) IsValid() bool {
	return u.IsIdValid() && u.IsNicknameValid()
}

func (u *User) EncryptPassword() error {
	// 密码不能为空
	if u.Password == "" {
		return errors.New("密码不能为空")
	}
	// 密码加密
	encryptedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), 12)
	if err != nil {
		return errors.New("密码加密失败")
	}
	// 密码赋值
	u.Password = string(encryptedPassword)
	return nil
}
