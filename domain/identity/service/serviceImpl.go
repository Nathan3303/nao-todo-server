package service

import (
	"context"
	"errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/domain/types"
)

func NewIdentityDomain(
	jwt repositories.JWT,
	userSessionRepo repositories.UserSession,
	rateLimitRepo repositories.RateLimit,
) IdentityDomain {
	return &identityDomainImpl{
		jwtRepo:         jwt,
		userSessionRepo: userSessionRepo,
		rateLimitRepo:   rateLimitRepo,
	}
}

// CreateSession 创建用户会话
func (d *identityDomainImpl) CreateSession(
	ctx context.Context,
	userId types.UserID,
	token string,
) error {
	return d.userSessionRepo.Create(
		ctx,
		&entities.UserSession{UserId: userId, Token: token},
	)
}

// FindSessionByUserIdAndToken 根据用户ID和会话令牌查找用户会话
func (d *identityDomainImpl) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId types.UserID,
	token string,
) (*entities.UserSession, error) {
	session := d.userSessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, errors.New("会话不存在")
	}
	return session, nil
}

// DeleteSession 删除用户会话
func (d *identityDomainImpl) DeleteSession(
	ctx context.Context,
	sessionEntity *entities.UserSession,
) error {
	return d.userSessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

// GenerateJWT 生成JWT令牌
func (d *identityDomainImpl) GenerateJWT(
	ctx context.Context,
	userEntity *entities.User,
) (string, error) {
	if userEntity.Id <= 0 {
		return "", errors.New("userEntity Id 无效")
	}
	jwtClaims := &valueobjects.JWTClaims{
		UserId: userEntity.Id,
		Email:  userEntity.Email,
	}
	return d.jwtRepo.Generate(ctx, jwtClaims)
}

// ParseJWT 解析JWT令牌
func (d *identityDomainImpl) ParseJWT(ctx context.Context, token string) (types.UserID, error) {
	jwtClaims, err := d.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return types.UserID(jwtClaims.UserId), nil
}

// CheckRateLimit 检查用户请求次数是否超过限流阈值
func (d *identityDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int8) error {
	if d.rateLimitRepo.Get(ctx, key) >= limit {
		return errors.New("用户请求次数超过限流阈值")
	}
	return d.rateLimitRepo.Incr(ctx, key)
}
