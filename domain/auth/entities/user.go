package entities

import (
	"time"
)

// User 用户实体
type User struct {
	Id          int64
	Account     string
	Email       string
	Password    string
	Nickname    string
	Avatar      string
	CreatedFrom string
	Role        string
	State       int8
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeactivedAt time.Time
}
