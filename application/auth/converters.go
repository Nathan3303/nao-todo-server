package auth

import (
	"naotodoserver/domain/auth/entities"
	"naotodoserver/interfaces/types"
)

func SignUpReqToUserEntity(signUpReq *types.UserSignUpReq) *entities.User {
	return &entities.User{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
		Nickname: signUpReq.Nickname,
	}
}
