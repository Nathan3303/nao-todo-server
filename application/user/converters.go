package user

import (
	"naotodoserver/domain/user/entities"
	"naotodoserver/interfaces/types"
)

func UserEntity2Res(e *entities.User) *types.GetUserProfileRes {
	res := &types.GetUserProfileRes{}
	res.Email = e.Email
	res.Nickname = e.Nickname
	res.Avatar = e.Avatar
	res.Role = e.Role
	res.CreatedFrom = e.CreatedFrom
	res.State = e.State
	res.Config = e.Config
	res.CreatedAt = e.CreatedAt.Format("2006-01-02 15:04:05")
	res.UpdatedAt = e.UpdatedAt.Format("2006-01-02 15:04:05")
	return res
}
