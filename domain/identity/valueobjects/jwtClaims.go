package valueobjects

import "time"

type JWTClaims struct {
	UserId    int64
	Email     string
	ExpOffset int64
	IssuedAt  time.Time
	Issuer    string
}
