package repositories

import (
	"context"
	"naotodoserver/domain/auth/valueobjects"
)

type JWT interface {
	Generate(ctx context.Context, jwtClaims *valueobjects.JWTClaims) (string, error)
	Validate(ctx context.Context, jwtString string) bool
	Parse(ctx context.Context, jwtString string) (*valueobjects.JWTClaims, error)
}
