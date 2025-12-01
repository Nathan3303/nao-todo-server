package auth

import (
	"naotodoserver/domain/auth/entities"
	"naotodoserver/interfaces/types"
)

func SignUpReqToUserEntity(signUpReq *types.SignUpReq) *entities.User {
	return &entities.User{
		Email:    signUpReq.Email,
		Password: signUpReq.Password,
	}
}
