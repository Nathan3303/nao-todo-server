package identity

import (
	"context"
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

// IsExpired 检查 JWT 是否已过期（解析失败视为已过期）
func (r *JWTRepoImpl) IsExpired(ctx context.Context, jwtString string) bool {
	return auth.GetJWTService().IsTokenExpired(jwtString)
}

func (r *JWTRepoImpl) Parse(
	ctx context.Context,
	jwtString string,
) (*valueobjects.JWTClaims, error) {
	// 仅验签与解析；过期判断由业务层负责（Validate 检查过期、CheckIn 依赖会话有效期）
	claims, err := auth.GetJWTService().Parse(jwtString)
	if err != nil {
		return nil, err
	}
	return Claims2JWTClaimsVO(claims), nil
}
