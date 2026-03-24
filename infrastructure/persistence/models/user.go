package models

import "database/sql"

type User struct {
	ModelBase
	Account     string       `gorm:"size:64;unique;not null"`
	Email       string       `gorm:"size:64;unique;not null"`
	Password    string       `gorm:"size:64;not null"`
	Nickname    string       `gorm:"size:32;not null"`
	Avatar      string       `gorm:"size:256;default:'https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif?imageView2/1/w/80/h/80'"`
	CreatedFrom string       `gorm:"size:64;default:'unknown:unknown'"`
	Role        string       `gorm:"size:32;default:'user'"`
	State       int8         `gorm:"default:1"`
	Config      *UserConfig  `gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:CASCADE;"`
	DeactivedAt sql.NullTime `gorm:"null"`
}

type UserConfig struct {
	ModelBase
	UserId     int64  `gorm:"not null"`
	State      string `gorm:"size:32;default:'active'"`
	Appearance string `gorm:"size:32;default:'light'"`
}
