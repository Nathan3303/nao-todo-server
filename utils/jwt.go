package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserJWTClaimsProfile struct {
	Id        int64     `json:"id"`
	Email     string    `json:"email"`
	Nickname  string    `json:"nickname"`
	Avatar    string    `json:"avatar"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

type UserJWTClaims struct {
	Profile UserJWTClaimsProfile `json:"profile"`
	jwt.RegisteredClaims
}

var JwtSecret = []byte("abc.key")

func GenerateUserJWT(profile UserJWTClaimsProfile) (string, error) {
	iUserJWTClaims := UserJWTClaims{
		Profile: UserJWTClaimsProfile{
			Id:        profile.Id,
			Email:     profile.Email,
			Nickname:  profile.Nickname,
			Avatar:    profile.Avatar,
			Role:      profile.Role,
			CreatedAt: profile.CreatedAt,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "Token",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, iUserJWTClaims)

	jwtString, err := token.SignedString(JwtSecret)
	if err != nil {
		return "", err
	}

	return jwtString, nil
}

func ParseUserJWT(tokenString string) (UserJWTClaims, error) {
	iUserJWTClaims := UserJWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &iUserJWTClaims, func(token *jwt.Token) (any, error) {
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return iUserJWTClaims, errors.New("非法的用户 JWT")
	}

	return iUserJWTClaims, nil
}

func IsUserJWTExpired(tokenString string) bool {
	iUserJWTClaims, err := ParseUserJWT(tokenString)
	if err != nil {
		return true
	}

	return IsUserJWTExpiredByClaims(&iUserJWTClaims)
}

func IsUserJWTExpiredByClaims(iUserJWTClaims *UserJWTClaims) bool {
	return iUserJWTClaims.RegisteredClaims.ExpiresAt.Before(time.Now())
}
