package user

import (
	"naotodoserver/domain/user/entities"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"
)

func UserEntity2Model(e *entities.User) *models.User {
	u := &models.User{}
	u.ID = e.Id
	u.Account = e.Account
	u.Email = e.Email
	u.Nickname = e.Nickname
	u.Password = e.Password
	u.Role = e.Role
	u.State = e.State
	u.CreatedAt = e.CreatedAt
	u.UpdatedAt = e.UpdatedAt
	u.DeletedAt = e.DeletedAt
	u.DeactivedAt = utils.TimePtr2SqlNullTime(e.DeactivedAt)
	return u
}

func UserModel2Entity(m *models.User) *entities.User {
	u := &entities.User{}
	u.Id = m.ID
	u.Account = m.Account
	u.Email = m.Email
	u.Nickname = m.Nickname
	u.Password = m.Password
	u.Avatar = m.Avatar
	u.Role = m.Role
	u.State = m.State
	u.CreatedFrom = m.CreatedFrom
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	u.DeletedAt = m.DeletedAt
	u.DeactivedAt = utils.SqlNullTime2TimePtr(m.DeactivedAt)
	return u
}

func UserConfigEntity2Model(e *entities.UserConfig) *models.UserConfig {
	u := &models.UserConfig{}
	u.ID = e.Id
	u.UserId = e.UserId
	u.State = e.State
	u.Appearance = e.Appearance
	u.CreatedAt = e.CreatedAt
	u.UpdatedAt = e.UpdatedAt
	u.DeletedAt = e.DeletedAt
	return u
}

func UserConfigModel2Entity(m *models.UserConfig) *entities.UserConfig {
	u := &entities.UserConfig{}
	u.Id = m.ID
	u.UserId = m.UserId
	u.State = m.State
	u.Appearance = m.Appearance
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	u.DeletedAt = m.DeletedAt
	return u
}
