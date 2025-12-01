package auth

import (
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/vo"
	"naotodoserver/infrastructure/auth"
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

func Claims2JWTClaimsVO(c *auth.Claims) *vo.JWTClaims {
	jc := &vo.JWTClaims{}
	jc.UserId = c.Id
	jc.Email = c.Payload
	jc.IssuedAt = c.IssuedAt.Time
	jc.Issuer = c.Issuer
	return jc
}
