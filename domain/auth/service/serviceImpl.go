package service

import (
	"context"
	"errors"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/repositories"
	"naotodoserver/domain/auth/vo"
)

func GetAuthDomainImpl(
	jwt repositories.JWT,
	userRepo repositories.User,
	sessionRepo repositories.Session,
	rateLimitRepo repositories.RateLimit,
) AuthDomain {
	return &authDomainImpl{
		jwtRepo:       jwt,
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		rateLimitRepo: rateLimitRepo,
	}
}

/*
 * Create User
 * 创建用户
 */
func (a *authDomainImpl) CreateUser(
	ctx context.Context,
	userEntity *entities.User,
) (*entities.User, error) {
	// 1. 密码加密
	err := userEntity.EncryptPassword()
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	// 2. 填充默认值
	userEntity.FillDefaultValue()
	// 3. 创建用户
	return a.userRepo.Create(ctx, userEntity)
}

/*
 * Create Session
 * 创建会话
 */
func (a *authDomainImpl) CreateSession(ctx context.Context, userId int64, token string) error {
	return a.sessionRepo.Create(ctx, &entities.Session{UserId: userId, Token: token})
}

/*
 * Find User By Email
 * 根据邮箱查找用户
 */
func (a *authDomainImpl) FindUserByEmail(
	ctx context.Context,
	email string,
) (*entities.User, error) {
	return a.userRepo.FindByEmail(ctx, email)
}

/*
 * Find User By Id
 * 根据 Id 查找用户
 */
func (a *authDomainImpl) FindUserById(
	ctx context.Context,
	id int64,
) (*entities.User, error) {
	return a.userRepo.FindById(ctx, id)
}

/*
 * Find Session By UserId And Token
 * 根据用户 Id 和 Token 查找会话
 */
func (a *authDomainImpl) FindSessionByUserIdAndToken(
	ctx context.Context,
	userId int64,
	token string,
) (*entities.Session, error) {
	session := a.sessionRepo.FindByUserIdAndToken(ctx, userId, token)
	if !session.IsValid() {
		return nil, errors.New("会话不存在")
	}
	return session, nil
}

/*
 * Update Session Token
 * 更新会话中 Token 字段值 (新的 JWT)
 */
func (a *authDomainImpl) UpdateSessionToken(
	ctx context.Context,
	sessionEntity *entities.Session,
) error {
	return a.sessionRepo.UpdateToken(ctx, sessionEntity)
}

/*
 * Delete Session By UserId And Token
 * 根据用户 Id 和 Token 删除会话
 */
func (a *authDomainImpl) DeleteSession(ctx context.Context, sessionEntity *entities.Session) error {
	return a.sessionRepo.Delete(ctx, sessionEntity.UserId, sessionEntity.Token)
}

/*
 * Generate JWT
 * 生成 JWT
 */
func (a *authDomainImpl) GenerateJWT(
	ctx context.Context,
	userEntity *entities.User,
) (string, error) {
	// 1. 判断 userEntity Id 是否合法
	if !userEntity.IsIdValid() {
		return "", errors.New("userEntity Id 无效")
	}
	// 2. 创建 JWTClaims
	jwtClaims := &vo.JWTClaims{
		UserId: userEntity.Id,
		Email:  userEntity.Email,
	}
	// 3. 创建 JWT
	token, err := a.jwtRepo.Generate(ctx, jwtClaims)
	if err != nil {
		return "", err
	}
	// 4. 返回结果
	return token, nil
}

/*
 * Parse JWT
 * 解析 JWT 并返回用户 Id - JWT 无效时会返回错误
 */
func (a *authDomainImpl) ParseJWT(ctx context.Context, token string) (int64, error) {
	jwtClaims, err := a.jwtRepo.Parse(ctx, token)
	if err != nil {
		return 0, err
	}
	return jwtClaims.UserId, nil
}

/*
 * Compare Passwords
 * 密码比较
 */
func (a *authDomainImpl) PasswordCompare(
	ctx context.Context,
	password []byte,
	encryptedPassword []byte,
) bool {
	return a.userRepo.PasswordCompare(password, encryptedPassword)
}

/*
 * Check Rate Limit
 * 检查用户请求次数是否超出阈值 - 避免暴力注册
 */
func (a *authDomainImpl) CheckRateLimit(ctx context.Context, key string, limit int8) error {
	// 1. 通过 Key 获取用户限流次数
	rateLimit := a.rateLimitRepo.Get(ctx, key)
	// 2. 判断用户请求次数是否超过限流阈值（1分钟内允许1次）
	if rateLimit >= limit {
		return errors.New("用户请求次数超过限流阈值")
	}
	// 3. 增加限流数
	a.rateLimitRepo.Incr(ctx, key)
	return nil
}
