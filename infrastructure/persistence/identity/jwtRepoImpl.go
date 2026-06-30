package identity

import (
	"context"
	"errors"
	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/auth"
	"time"
)

type JWTRepoImpl struct{}

func NewJWTRepo() repositories.JWT {
	return &JWTRepoImpl{}
}

func (r *JWTRepoImpl) Generate(
	ctx context.Context,
	jwtClaims *valueobjects.JWTClaims,
) (string, error) {
	token, err := auth.GetJWTService().Generate(
		jwtClaims.UserId,
		jwtClaims.Email,
		time.Duration(jwtClaims.ExpOffset),
	)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *JWTRepoImpl) Validate(ctx context.Context, jwtString string) bool {
	return auth.GetJWTService().IsTokenExpired(jwtString)
}

func (r *JWTRepoImpl) Parse(
	ctx context.Context,
	jwtString string,
) (*valueobjects.JWTClaims, error) {
	claims, err := auth.GetJWTService().Parse(jwtString)
	if err != nil {
		return nil, err
	}
	if auth.GetJWTService().IsExpired(claims) {
		return nil, errors.New("token is expired")
	}
	return Claims2JWTClaimsVO(claims), nil
}
