package auth

import (
	"naotodoserver/application/auth/dto"
	"naotodoserver/domain/identity/valueobjects"
)

// SignUpInputToCreateUserValueObject 将 SignUpInput 转换为 CreateUserValueObject
// @param signUpInput 注册入参
// @return CreateUserValueObject 创建用户值对象
// @error 错误信息
func SignUpInputToCreateUserValueObject(
	signUpInput *dto.SignUpInput,
) (*valueobjects.CreateUser, error) {
	return valueobjects.NewCreateUser(
		signUpInput.Email,
		signUpInput.Password,
		signUpInput.Email,
		signUpInput.Nickname,
	)
}
