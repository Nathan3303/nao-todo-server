package vo

import "time"

type JWTClaims struct {
	UserId    int64     `json:"userId"`
	Email     string    `json:"email"`
	ExpOffset int64     `json:"expOffset"`
	IssuedAt  time.Time `json:"issuedAt"`
	Issuer    string    `json:"issuer"`
}
