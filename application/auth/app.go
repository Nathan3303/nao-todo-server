package auth

import (
	"context"
	"errors"
	"naotodoserver/domain/auth/entities"
	"naotodoserver/domain/auth/service"
	"naotodoserver/interfaces/types"
	"sync"
)

type AuthAppBC interface {
	SignIn(ctx context.Context, signInReq types.UserSignInReq) (*types.UserSignInRes, error)
	SignUp(ctx context.Context, signUpReq types.UserSignUpReq) (*types.UserSignUpRes, error)
	SignOut(ctx context.Context, signOutReq types.UserSignOutReq) (*types.UserSignOutRes, error)
	CheckIn(ctx context.Context, checkInReq types.UserCheckInReq) (*types.UserCheckInRes, error)
}

type AuthApp struct {
	authDomain *service.AuthDomain
}

var (
	authAppImpl *AuthApp
	once        sync.Once
)

func RegistDomain(authDomain *service.AuthDomain) AuthAppBC {
	once.Do(func() {
		authAppImpl = &AuthApp{authDomain: authDomain}
	})
	return authAppImpl
}

/**
 * SignIn 处理用户登录
 * 通过 signInReq 中的邮箱和密码验证用户身份，验证成功后生成 JWT 令牌并创建用户会话，最后返回 JWT 令牌
 */
func (as *AuthApp) SignIn(
	ctx context.Context,
	signInReq types.UserSignInReq,
) (*types.UserSignInRes, error) {
	// 1. 通过 Email 查找用户记录
	user, err := as.authDomain.FindUserByEmail(ctx, signInReq.Email)
	if err != nil || !user.IsValid() {
		return nil, errors.New("用户名或密码错误")
	}
	// 2. 比对密码
	isMatched := as.authDomain.PasswordCompare(
		ctx,
		[]byte(signInReq.Password),
		[]byte(user.Password),
	)
	if !isMatched {
		return nil, errors.New("用户名或密码错误")
	}
	// 2. 生成 JWT
	jwtString, err := as.authDomain.GenerateJWT(ctx, user)
	if err != nil {
		return nil, errors.New("用户凭证签发失败")
	}
	// 3. 创建 Session
	err = as.authDomain.CreateSession(ctx, user.Id, jwtString)
	if err != nil {
		return nil, errors.New("用户 Session 创建失败 - " + err.Error())
	}
	// 4. 登录成功 返回 JWT
	return &types.UserSignInRes{Token: jwtString}, nil
}

/**
 * SignUp 处理用户注册
 * 通过 signUpReq 中的用户信息，执行密码加密后用户信息落库
 */
func (as *AuthApp) SignUp(
	ctx context.Context,
	signUpReq types.UserSignUpReq,
) (*types.UserSignUpRes, error) {
	// 1. 查找用户 并 检测查找结果是否存在
	userEntity, _ := as.authDomain.FindUserByEmail(ctx, signUpReq.Email)
	if userEntity.IsValid() {
		return nil, errors.New("用户已存在")
	}
	// 2. 更新用户实体 - 密码加密
	userEntity = SignUpReqToUserEntity(&signUpReq)
	err := userEntity.EncryptPassword()
	if err != nil {
		return nil, errors.New("密码加密失败")
	}
	// 3. 创建用户
	_, err = as.authDomain.CreateUser(ctx, userEntity)
	if err != nil {
		return nil, errors.New("注册用户失败")
	}
	// 4. 注册成功
	return &types.UserSignUpRes{}, nil
}

/**
 * CheckIn 处理用户检入
 * 通过 checkInReq 中的令牌信息，验证用户会话，并生成新的 JWT 并更新用户会话，最后返回新的 JWT
 * 用于在 JWT 或会话期限内的免密登录
 */
func (as *AuthApp) CheckIn(
	ctx context.Context,
	checkInReq types.UserCheckInReq,
) (*types.UserCheckInRes, error) {
	// // 1. 校验 JWT
	// fmt.Println("checkInReq.Token:", checkInReq.Token)
	// err, userJWTPayload := s.UserDomain.ValidateJWT(ctx, checkInReq.Token)
	// if err != nil {
	// 	return nil, errors.New("用户凭证验证失败 - " + err.Error())
	// }
	// // 2. 验证 Session 会话
	// isSessionValid, userSession := s.UserDomain.ValidateSession(ctx, checkInReq.Token)
	// if !isSessionValid {
	// 	return nil, errors.New("用户会话验证失败")
	// }
	// // 3. JWT 转换实体
	// userEntity, ok := vo.UserJWTPayloadToUserEntity(userJWTPayload)
	// if !ok {
	// 	return nil, errors.New("用户信息转换失败")
	// }
	// // 3. 重新生成 JWT
	// newJWT, err := s.UserDomain.GenerateJWT(ctx, userEntity)
	// if err != nil {
	// 	return nil, errors.New("用户凭证签发失败 - " + err.Error())
	// }
	// // 4. 更新 Session
	// err = s.UserDomain.UpdateSessionToken(ctx, userSession, newJWT)
	// if err != nil {
	// 	return nil, errors.New("用户 Session 更新失败 - " + err.Error())
	// }
	// // 4. 返回用户信息
	// return &types.UserCheckInRes{Token: newJWT}, nil
	panic("unimplemented")
}

/**
 * SignOut 处理用户登出
 * 通过 signOutReq 中的令牌信息，验证用户会话，并删除用户会话，最后返回成功结果
 */
func (as *AuthApp) SignOut(
	ctx context.Context,
	signOutReq types.UserSignOutReq,
) (*types.UserSignOutRes, error) {
	// 1. 验证 JWT 并处理失败结果
	isValid := as.authDomain.ValidateJWT(ctx, signOutReq.Token)
	if isValid {
		return nil, errors.New("用户凭证验证失败")
	}
	// 2. 删除 Session
	err := as.authDomain.DeleteSession(ctx, &entities.Session{Token: signOutReq.Token})
	if err != nil {
		return nil, errors.New("用户 Session 删除失败 - " + err.Error())
	}
	// 3. 返回结果
	return &types.UserSignOutRes{}, nil
}
