package repositories

import (
	"context"
	"naotodoserver/domain/auth/vo"
)

type JWT interface {
	Generate(ctx context.Context, jwtClaims *vo.JWTClaims) (string, error)
	Validate(ctx context.Context, jwtString string) bool
	Parse(ctx context.Context, jwtString string) (*vo.JWTClaims, error)
}
