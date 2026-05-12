package models

import "database/sql"

type User struct {
	ModelBase
	Account     string       `gorm:"size:64;unique;not null;index:idx_user_account"`
	Email       string       `gorm:"size:64;unique;not null;index:idx_user_email"`
	Password    string       `gorm:"size:64;not null"`
	Nickname    string       `gorm:"size:64;not null"`
	Avatar      string       `gorm:"size:256;default:'https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif?imageView2/1/w/80/h/80'"`
	CreatedFrom string       `gorm:"size:64;default:'unknown:unknown'"`
	Role        string       `gorm:"size:32;default:'user';index:idx_user_role"`
	State       int8         `gorm:"default:1;index:idx_user_state"`
	Config      *UserConfig  `gorm:"foreignKey:UserId;references:ID;constraint:OnDelete:CASCADE;"`
	DeactivedAt sql.NullTime `gorm:"null;index:idx_user_deactived_at"`
}

type UserConfig struct {
	ModelBase
	UserId     int64  `gorm:"not null;index:idx_user_config_user"`
	Appearance string `gorm:"size:32;default:'auto'"`
}
