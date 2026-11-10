package identity

import (
	"context"
	"fmt"
	"time"

	"naotodoserver/domain/identity/repositories"
	"naotodoserver/domain/identity/valueobjects"
	"naotodoserver/infrastructure/auth"
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
	// 断言字段残缺（如缺 iat / iat 为 null）显式报错：调用方已把解析错误归为凭证类，
	// 不得解引用 nil 指针 panic（DEF-157）
	claimsVO, err := Claims2JWTClaimsVO(claims)
	if err != nil {
		return nil, fmt.Errorf("jwtRepo.Parse: %w", err)
	}
	return claimsVO, nil
}
