package auth

import (
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/valueobjects"
	"naotodoserver/infrastructure/auth"
	"naotodoserver/infrastructure/persistence/models"
)

func UserEntity2Model(e *entities.User) *models.User {
	u := &models.User{}
	u.ID = e.Id
	u.Account = e.Account
	u.Email = e.Email
	u.Password = e.Password
	u.Nickname = e.Nickname
	u.Avatar = e.Avatar
	u.CreatedFrom = e.CreatedFrom
	u.Role = e.Role
	u.CreatedAt = e.CreatedAt
	u.UpdatedAt = e.UpdatedAt
	return u
}

func UserModel2Entity(m *models.User) *entities.User {
	u := &entities.User{}
	u.Id = m.ID
	u.Account = m.Account
	u.Email = m.Email
	u.Password = m.Password
	u.Nickname = m.Nickname
	u.Avatar = m.Avatar
	u.CreatedFrom = m.CreatedFrom
	u.Role = m.Role
	u.CreatedAt = m.CreatedAt
	u.UpdatedAt = m.UpdatedAt
	return u
}

func SessionModel2Entity(m *models.Session) *entities.Session {
	s := &entities.Session{}
	s.Id = m.ID
	s.UserId = m.UserId
	s.Token = m.Token
	s.ExpiredAt = m.ExpiredAt
	s.DeviceType = m.DeviceType
	return s
}

func SessionEntity2Model(e *entities.Session) *models.Session {
	s := &models.Session{}
	s.ID = e.Id
	s.UserId = e.UserId
	s.Token = e.Token
	s.ExpiredAt = e.ExpiredAt
	s.DeviceType = e.DeviceType
	return s
}

func Claims2JWTClaimsVO(c *auth.Claims) *valueobjects.JWTClaims {
	jc := &valueobjects.JWTClaims{}
	jc.UserId = c.Id
	jc.Email = c.Payload
	jc.IssuedAt = c.IssuedAt.Time
	jc.Issuer = c.Issuer
	return jc
}

func CreateUserValueObjectToModel(createUserValueObject *valueobjects.CreateUser) *models.User {
	return &models.User{
		Account:  createUserValueObject.Email,
		Email:    createUserValueObject.Email,
		Password: createUserValueObject.EncryptedPassword,
		Nickname: createUserValueObject.Nickname,
	}
}
