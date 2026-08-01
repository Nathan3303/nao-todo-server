package user

import (
	"naotodoserver/application/user/dto"
	"naotodoserver/conf"
	"naotodoserver/domain/identity/entities"
	"time"
)

// UserEntity2Res 用户实体转换为获取用户响应
// @param e 用户实体
// @return 获取用户响应
func UserEntity2Res(e *entities.User) *dto.GetProfileOutput {
	res := &dto.GetProfileOutput{}
	res.Email = e.Email
	res.Nickname = e.Nickname
	res.Avatar = conf.Conf.Uploads.AvatarURL(e.Avatar)
	res.Role = ""
	res.CreatedFrom = e.CreatedFrom
	res.State = uint8(e.State)
	if deactivedAt, ok := e.DeactivedAt.Value(); ok {
		res.DeactivedAt = deactivedAt.Add(time.Hour * 24 * 7).Format(time.RFC3339)
	}
	if lastRestoreAt, ok := e.LastCancelRestoreAt.Value(); ok {
		res.LastRestoreAt = lastRestoreAt.Format(time.RFC3339)
	}
	res.Config = e.Config
	res.CreatedAt = e.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = e.UpdatedAt.Format(time.RFC3339)
	return res
}

// ConfigEntity2Res 用户配置实体转换为获取用户配置响应
func ConfigEntity2Res(e *entities.UserConfig) *dto.GetConfigOutput {
	return &dto.GetConfigOutput{
		Appearance: e.Appearance,
	}
}
