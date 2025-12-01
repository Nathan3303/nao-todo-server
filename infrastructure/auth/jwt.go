package auth

import (
	"errors"
	"naotodoserver/conf"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	Id      int64
	Payload string
	jwt.RegisteredClaims
}

type JWTService interface {
	Generate(id int64, payload string, expOffset time.Duration) (string, error)
	Parse(jwtString string) (Claims, error)
	IsTokenExpired(jwtString string) bool
	IsExpired(claims *Claims) bool
}

type JWTServiceImpl struct {
	Secret []byte
}

var (
	jwtService *JWTServiceImpl
	once       sync.Once
)

func GetJWTService() *JWTServiceImpl {
	once.Do(func() {
		jwtService = &JWTServiceImpl{
			Secret: []byte(conf.Conf.Server.JwtSecret),
		}
	})
	return jwtService
}

func (jwtService *JWTServiceImpl) Generate(
	id int64,
	payload string,
	expOffset time.Duration,
) (string, error) {
	if expOffset == 0 {
		expOffset = time.Hour * 48
	}
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, iClaims)

	jwtString, err := token.SignedString(jwtService.Secret)
	if err != nil {
		return "", err
	}

	return jwtString, nil
}

func (jwtService *JWTServiceImpl) Parse(jwtString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		jwtString,
		claims,
		func(token *jwt.Token) (any, error) {
			return jwtService.Secret, nil
		},
	)
	if err != nil || !token.Valid {
		return claims, errors.New("非法的用户 JWT")
	}

	return claims, nil
}

func (jwtService *JWTServiceImpl) IsTokenExpired(jwtString string) bool {
	claims, err := jwtService.Parse(jwtString)
	if err != nil {
		return true
	}

	return jwtService.IsExpired(claims)
}

func (jwtService *JWTServiceImpl) IsExpired(claims *Claims) bool {
	if claims == nil || claims.RegisteredClaims.ExpiresAt == nil {
		return true
	}
	return claims.RegisteredClaims.ExpiresAt.Before(time.Now())
}
