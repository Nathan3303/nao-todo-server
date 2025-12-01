package auth

import (
	"naotodoserver/domain/auth/entities"
	"naotodoserver/infrastructure/persistence/models"
)

func UserEntity2Model(e *entities.User) *models.User {
	u := &models.User{}
	u.ID = e.Id
	u.Account = e.Account
	u.Email = e.Email
	u.Nickname = e.Nickname
	u.Password = e.Password
	return u
}

func UserModel2Entity(m *models.User) *entities.User {
	u := &entities.User{}
	u.Id = m.ID
	u.Account = m.Account
	u.Email = m.Email
	u.Nickname = m.Nickname
	u.Password = m.Password
	return u
}
