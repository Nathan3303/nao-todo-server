package identity

import (
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/auth"
	"naotodoserver/infrastructure/persistence/models"
	"naotodoserver/infrastructure/utils"

	"gorm.io/gorm"
)

// === User converters ===

func UserEntity2Model(e *entities.User) *models.User {
	u := &models.User{
		Account:     e.Account,
		Email:       e.Email,
		Password:    e.Password,
		Nickname:    e.Nickname,
		Avatar:      e.Avatar,
		CreatedFrom: e.CreatedFrom,
		Role:        e.Role,
		State:       e.State,
		DeactivedAt: utils.TimePtr2SqlNullTime(e.DeactivedAt),
	}
	u.ID = e.Id
	u.CreatedAt = e.CreatedAt
	u.UpdatedAt = e.UpdatedAt
	u.DeletedAt = gorm.DeletedAt{Time: e.DeletedAt, Valid: !e.DeletedAt.IsZero()}
	return u
}

func UserModel2Entity(m *models.User) *entities.User {
	return &entities.User{
		Id:          m.ID,
		Account:     m.Account,
		Email:       m.Email,
		Password:    m.Password,
		Nickname:    m.Nickname,
		Avatar:      m.Avatar,
		CreatedFrom: m.CreatedFrom,
		Role:        m.Role,
		State:       m.State,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt.Time,
		DeactivedAt: utils.SqlNullTime2TimePtr(m.DeactivedAt),
	}
}

// === UserConfig converters ===

func UserConfigEntity2Model(e *entities.UserConfig) *models.UserConfig {
	return &models.UserConfig{
		UserId:     e.UserId,
		Appearance: e.Appearance,
	}
}

func UserConfigModel2Entity(m *models.UserConfig) *entities.UserConfig {
	return &entities.UserConfig{
		Id:         m.ID,
		UserId:     m.UserId,
		Appearance: m.Appearance,
		CreatedAt:  m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
		DeletedAt:  m.DeletedAt.Time,
	}
}

// === Session converters ===

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

// === JWT converters ===

func Claims2JWTClaimsVO(c *auth.Claims) *valueobjects.JWTClaims {
	return &valueobjects.JWTClaims{
		UserId:    c.Id,
		Email:     c.Payload,
		IssuedAt:  c.IssuedAt.Time,
		Issuer:    c.Issuer,
	}
}
