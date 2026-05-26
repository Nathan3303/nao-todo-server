package auth

import (
	"errors"
	"naotodoserver/conf"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims JWT 断言结构体
// @param Id 用户 ID
// @param Payload 有效载荷
// @param RegisteredClaims 注册的断言
type Claims struct {
	Id      int64
	Payload string
	jwt.RegisteredClaims
}

// JWTService JWT 服务接口
// @param Generate 生成 JWT
// @param Parse 解析 JWT
// @param IsTokenExpired 检查 JWT 是否过期
// @param IsExpired 检查 JWT 断言是否过期
type JWTService interface {
	Generate(id int64, payload string, expOffset time.Duration) (string, error)
	Parse(jwtString string) (*Claims, error)
	IsTokenExpired(jwtString string) bool
	IsExpired(claims *Claims) bool
}

// JWTServiceImpl JWT 服务实现
// @param Secret JWT 密钥
type JWTServiceImpl struct {
	Secret []byte
}

// jwtService JWT 服务实例例
var (
	jwtService *JWTServiceImpl
	once       sync.Once
)

// GetJWTService 获取 JWT 服务
// @return JWT 服务
// @note 该服务为单例模式，仅在应用启动时初始化一次
func GetJWTService() *JWTServiceImpl {
	once.Do(func() {
		jwtService = &JWTServiceImpl{
			Secret: []byte(conf.Conf.Server.JwtSecret),
		}
	})
	return jwtService
}

/**
 * 生成 JWT
 * @param id 用户 ID
 * @param payload 有效载荷
 * @param expOffset 过期时间偏移量，默认 48 小时
 * @return JWT 字符串
 * @note 该方法会根据提供的参数生成一个 JWT 字符串，用于标识用户身份和传递信息
 */
func (jwtService *JWTServiceImpl) Generate(
	id int64,
	payload string,
	expOffset time.Duration,
) (string, error) {
	// 如果未指定过期时间偏移量，默认 48 小时
	if expOffset == 0 {
		expOffset = time.Hour * 48
	}
	// 构建 JWT 断言
	iClaims := Claims{
		Id:      id,
		Payload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "Token",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expOffset)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "NaoTodoServer",
		},
	}
	// 构建 JWT 令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, iClaims)
	// 签名 JWT 令牌
	jwtString, err := token.SignedString(jwtService.Secret)
	if err != nil {
		return "", err
	}
	return jwtString, nil
}

/**
 * 解析 JWT
 * @param jwtString JWT 字符串
 * @return 解析后的 JWT 断言
 * @note 该方法会尝试解析提供的 JWT 字符串，返回解析后的断言对象
 */
func (jwtService *JWTServiceImpl) Parse(jwtString string) (*Claims, error) {
	claims := &Claims{}
	// 解析 JWT 字符串
	token, err := jwt.ParseWithClaims(
		jwtString,
		claims,
		func(token *jwt.Token) (any, error) {
			return jwtService.Secret, nil
		},
	)
	// 检查解析结果是否有效
	if err != nil || !token.Valid {
		return claims, errors.New("非法的用户 JWT")
	}
	return claims, nil
}

/**
 * 检查 JWT 是否过期
 * @param jwtString JWT 字符串
 * @return 是否过期
 * @note 该方法会尝试解析提供的 JWT 字符串，检查其是否过期
 */
func (jwtService *JWTServiceImpl) IsTokenExpired(jwtString string) bool {
	// 解析 JWT 字符串
	claims, err := jwtService.Parse(jwtString)
	// 检查解析结果是否有效
	if err != nil {
		return true
	}
	return jwtService.IsExpired(claims)
}

/**
 * 检查 JWT 断言是否过期
 * @param claims JWT 断言
 * @return 是否过期
 * @note 该方法会检查提供的 JWT 断言是否过期，返回是否过期的结果
 */
func (jwtService *JWTServiceImpl) IsExpired(claims *Claims) bool {
	// 检查断言是否为空或过期时间是否为空
	if claims == nil || claims.RegisteredClaims.ExpiresAt == nil {
		return true
	}
	// 检查过期时间是否已过期
	return claims.RegisteredClaims.ExpiresAt.Before(time.Now())
}
