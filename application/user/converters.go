package user

import (
	"naotodoserver/domain/identity/entities"
	"naotodoserver/interfaces/types"
	"time"
)

// UserEntity2Res 用户实体转换为获取用户响应
// @param e 用户实体
// @return 获取用户响应
func UserEntity2Res(e *entities.User) *types.GetUserProfileRes {
	res := &types.GetUserProfileRes{}
	res.Email = e.Email
	res.Nickname = e.Nickname
	res.Avatar = e.Avatar
	res.Role = ""
	res.CreatedFrom = e.CreatedFrom
	res.State = e.State
	res.Config = e.Config
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	return res
}

// ConfigEntity2Res 用户配置实体转换为获取用户配置响应
func ConfigEntity2Res(e *entities.UserConfig) *types.GetUserConfigRes {
	return &types.GetUserConfigRes{
		Appearance: e.Appearance,
	}
}
