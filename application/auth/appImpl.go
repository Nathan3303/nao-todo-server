package auth

import (
	"context"
	"errors"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/service"
	"naotodoserver/interfaces/types"
)

// NewAuthApp 创建认证应用层实例
func NewAuthApp(identityDomain service.IdentityDomain, userRepo repositories.User, sessionRepo repositories.UserSession) AuthApp {
	return &authAppImpl{
		identityDomain: identityDomain,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
	}
}

// 处理用户登录
// @description 通过 signInReq 中的邮箱和密码验证用户身份，验证成功后生成 JWT 令牌并创建用户会话，最后返回 JWT 令牌
// @param ctx 上下文
// @param signInReq 登录请求
// @return 登录响应
// @return error 错误
func (as *authAppImpl) SignIn(
	ctx context.Context,
	signInReq *types.SignInReq,
) (*types.SignInRes, error) {
	// 1. 通过 Email 查找用户记录
	userEntity, err := as.userRepo.FindByEmail(ctx, signInReq.Email)
	if err != nil {
		return nil, err
	}
	// 2. 比对密码
	isMatched := as.userRepo.PasswordCompare(
		[]byte(signInReq.Password),
		[]byte(userEntity.Password),
	)
	if !isMatched {
		return nil, errors.New("用户名或密码错误")
	}
	// 2. 创建 JWT 令牌
	jwtString, err := as.identityDomain.GenerateJWT(ctx, userEntity)
	if err != nil {
		return nil, errors.New("用户凭证签发失败")
	}
	// 3. 创建 Session
	err = as.identityDomain.CreateSession(ctx, userEntity.Id, jwtString)
	if err != nil {
		return nil, errors.New("用户 Session 创建失败 - " + err.Error())
	}
	// 4. 登录成功 返回 JWT
	return &types.SignInRes{Token: jwtString}, nil
}

/*
 * SignUp 处理用户注册
 * 通过 signUpReq 中的用户信息，执行密码加密后用户信息落库
 */
func (as *authAppImpl) SignUp(
	ctx context.Context,
	signUpReq *types.SignUpReq,
) error {
	// 通过 Email 查找用户记录
	_, err := as.userRepo.FindByEmail(ctx, signUpReq.Email)
	if err == nil {
		return errors.New("邮箱已存在")
	}
	// 转换为 CreateUserValueObject
	createUserValueObject, err := SignUpReqToCreateUserValueObject(signUpReq)
	if err != nil {
		return err
	}
	// 创建
	_, err = as.userRepo.CreateByVO(ctx, createUserValueObject)
	if err != nil {
		return errors.New("注册用户失败")
	}
	// 注册成功
	return nil
}

/*
 * CheckIn 处理用户检入
 * 通过 checkInReq 中的令牌信息，验证用户会话，并生成新的 JWT 并更新用户会话，最后返回新的 JWT
 * 用于在 JWT 或会话期限内的免密登录
 */
func (as *authAppImpl) CheckIn(
	ctx context.Context,
	checkInReq *types.CheckInReq,
) (*types.CheckInRes, error) {
	// 1. 解析 JWT 令牌
	userId, err := as.identityDomain.ParseJWT(ctx, checkInReq.Token)
	if err != nil {
		return nil, errors.New("用户凭证验证失败 - " + err.Error())
	}
	// 2. 通过 JWT 令牌和用户 ID 查找会话
	sessionEntity, err := as.identityDomain.FindSessionByUserIdAndToken(
		ctx,
		userId,
		checkInReq.Token,
	)
	// 会话不存在：
	if err != nil {
		return nil, errors.New("用户会话验证失败 - " + err.Error())
	}
	if !sessionEntity.IsValid() {
		return nil, errors.New("用户会话验证失败")
	}
	// 会话存在：
	// 3. 查找最新的用户信息
	userEntity, err := as.userRepo.FindById(ctx, userId)
	if err != nil {
		return nil, errors.New("用户信息查询失败 - " + err.Error())
	}
	// 4. 构建新的 JWT 令牌
	newJWT, err := as.identityDomain.GenerateJWT(ctx, userEntity)
	if err != nil {
		return nil, errors.New("用户凭证签发失败 - " + err.Error())
	}
	// 5. 更新会话中的 Token 字段
	sessionEntity.Token = newJWT
	err = as.sessionRepo.UpdateToken(ctx, sessionEntity)
	if err != nil {
		return nil, errors.New("用户 Session 更新失败 - " + err.Error())
	}
	// 6. 返回
	return &types.CheckInRes{Token: newJWT}, nil
}

/*
 * SignOut 处理用户登出
 * 通过 signOutReq 中的令牌信息，验证用户会话，并删除用户会话，最后返回成功结果
 */
func (as *authAppImpl) SignOut(
	ctx context.Context,
	signOutReq *types.SignOutReq,
) error {
	// 1. 解析 JWT 令牌
	userId, err := as.identityDomain.ParseJWT(ctx, signOutReq.Token)
	if err != nil {
		return err
	}
	// 2. 通过 JWT 令牌和用户 ID 删除会话
	err = as.identityDomain.DeleteSession(ctx, &entities.UserSession{
		UserId: userId,
		Token:  signOutReq.Token,
	})
	if err != nil {
		return errors.New(err.Error())
	}
	// 3. 返回结果
	return nil
}

/*
 * Validate 处理用户令牌验证
 * 通过 Header 中的令牌信息，验证用户会话，并返回验证结果
 */
func (as *authAppImpl) Validate(
	ctx context.Context,
	token string,
) (int64, error) {
	// 1. 解析 JWT
	userId, err := as.identityDomain.ParseJWT(ctx, token)
	if err != nil {
		return 0, err
	}
	// 2. 检查会话
	session, err := as.identityDomain.FindSessionByUserIdAndToken(ctx, userId, token)
	if err != nil {
		return 0, err
	}
	if !session.IsValid() {
		return 0, errors.New("用户会话验证失败")
	}
	// 3. 返回结果
	return userId, nil
}

/*
 * RateLimit 处理用户限流
 * 通过 clientIP 检查用户请求次数是否超过限流阈值
 */
func (as *authAppImpl) RateLimit(
	ctx context.Context,
	clientIP string,
	limit int8,
) error {
	// 1. 构造限流 key
	key := "rate_limit:" + clientIP
	// 2. 检查是否超过限流阈值
	err := as.identityDomain.CheckRateLimit(ctx, key, limit)
	// 3. 返回结果
	return err
}
