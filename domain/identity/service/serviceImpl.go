package service

import (
	"context"
	"errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
)

func NewIdentityDomain(
	jwt repositories.JWT,
	userRepo repositories.User,
	sessionRepo repositories.Session,
	rateLimitRepo repositories.RateLimit,
) IdentityDomain {
	return &identityDomainImpl{
		jwtRepo:       jwt,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		rateLimitRepo: rateLimitRepo,
	}
}

// === 认证相关 ===

func (d *identityDomainImpl) CreateUser(ctx context.Context, vo *valueobjects.CreateUser) (*entities.User, error) {
	return d.userRepo.CreateByVO(ctx, vo)
}

func (d *identityDomainImpl) CreateSession(ctx context.Context, userId int64, token string) error {
	return d.sessionRepo.Create(ctx, &entities.Session{UserId: userId, Token: token})
}

func (d *identityDomainImpl) FindSessionByUserIdAndToken(ctx context.Context, userId int64, token string) (*entities.Session, error) {
	session := d.sessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, errors.New("会话不存在")
	}
	return session, nil
}

func (d *identityDomainImpl) UpdateSessionToken(ctx context.Context, sessionEntity *entities.Session) error {
	return d.sessionRepo.UpdateToken(ctx, sessionEntity)
}

func (d *identityDomainImpl) DeleteSession(ctx context.Context, sessionEntity *entities.Session) error {
	return d.sessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

func (d *identityDomainImpl) GenerateJWT(ctx context.Context, userEntity *entities.User) (string, error) {
	if userEntity.Id <= 0 {
		return "", errors.New("userEntity Id 无效")
	}
	jwtClaims := &valueobjects.JWTClaims{
		UserId: userEntity.Id,
		Email:  userEntity.Email,
	}
	return d.jwtRepo.Generate(ctx, jwtClaims)
}

func (d *identityDomainImpl) ParseJWT(ctx context.Context, token string) (int64, error) {
	jwtClaims, err := d.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return jwtClaims.UserId, nil
}

func (d *identityDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int8) error {
	if d.rateLimitRepo.Get(ctx, key) >= limit {
		return errors.New("用户请求次数超过限流阈值")
	}
	d.rateLimitRepo.Incr(ctx, key)
	return nil
}

// === 用户管理 ===

func (d *identityDomainImpl) FindByEmail(ctx context.Context, email string) (*entities.User, error) {
	return d.userRepo.FindByEmail(ctx, email)
}

func (d *identityDomainImpl) FindById(ctx context.Context, id int64) (*entities.User, error) {
	return d.userRepo.FindById(ctx, id)
}

func (d *identityDomainImpl) UpdateNickname(ctx context.Context, userId int64, nickname string) error {
	return d.userRepo.UpdateNickname(ctx, userId, nickname)
}

func (d *identityDomainImpl) UpdateAvatar(ctx context.Context, userId int64, avatar string) error {
	return d.userRepo.UpdateAvatar(ctx, userId, avatar)
}

func (d *identityDomainImpl) UpdatePassword(ctx context.Context, userId int64, password, newPassword string) error {
	return d.userRepo.UpdatePassword(ctx, userId, password, newPassword)
}

func (d *identityDomainImpl) Deactive(ctx context.Context, userId int64) error {
	return d.userRepo.Deactive(ctx, userId)
}

func (d *identityDomainImpl) Active(ctx context.Context, userId int64) error {
	return d.userRepo.Active(ctx, userId)
}

func (d *identityDomainImpl) PasswordCompare(password, encryptedPassword []byte) bool {
	return d.userRepo.PasswordCompare(password, encryptedPassword)
}

func (d *identityDomainImpl) GetConfig(ctx context.Context, userId int64) (*entities.UserConfig, error) {
	return d.userRepo.GetConfig(ctx, userId)
}

func (d *identityDomainImpl) UpdateConfig(ctx context.Context, userId int64, appearance string) error {
	return d.userRepo.UpdateConfig(ctx, userId, appearance)
}

func (d *identityDomainImpl) DeleteDeactivatedUsers(ctx context.Context, dayOffset int8) (int64, error) {
	return d.userRepo.DeleteDeactivatedUsers(ctx, dayOffset)
}
