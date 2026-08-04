package repositories

import (
	"context"
	"naotodoserver/domain/identity/valueobjects"
)

type JWT interface {
	Generate(ctx context.Context, jwtClaims *valueobjects.JWTClaims) (string, error)
	IsExpired(ctx context.Context, jwtString string) bool
	Parse(ctx context.Context, jwtString string) (*valueobjects.JWTClaims, error)
}
