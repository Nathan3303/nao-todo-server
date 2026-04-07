package auth

import (
	"naotodoserver/domain/auth/valueobjects"
	"naotodoserver/interfaces/types"
)

// 将 SignUpReq 转换为 CreateUserValueObject
// @param signUpReq 注册请求
// @return CreateUserValueObject 创建用户值对象
// @error 错误信息
func SignUpReqToCreateUserValueObject(
	signUpReq *types.SignUpReq,
) (*valueobjects.CreateUser, error) {
	return valueobjects.NewCreateUser(
		signUpReq.Email,
		signUpReq.Password,
		signUpReq.Email,
		signUpReq.Nickname,
	)
}
