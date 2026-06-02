package auth

import (
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/auth"
	"naotodoserver/infrastructure/persistence/models"
)

func SessionModel2Entity(m *models.Session) *entities.Session {
	return &entities.Session{
		Id:         m.ID,
		UserId:     m.UserId,
		Token:      m.Token,
		ExpiredAt:  m.ExpiredAt,
		DeviceType: m.DeviceType,
	}
}

func SessionEntity2Model(e *entities.Session) *models.Session {
	return &models.Session{
		UserId:     e.UserId,
		Token:      e.Token,
		ExpiredAt:  e.ExpiredAt,
		DeviceType: e.DeviceType,
	}
}

func Claims2JWTClaimsVO(c *auth.Claims) *valueobjects.JWTClaims {
	return &valueobjects.JWTClaims{
		UserId:    c.Id,
		Email:     c.Payload,
		IssuedAt:  c.IssuedAt.Time,
		Issuer:    c.Issuer,
	}
}
