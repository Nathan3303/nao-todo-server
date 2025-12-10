package models

type User struct {
	ModelBase
	Account    string      `gorm:"size:64 unique"`
	Email      string      `gorm:"size:64 unique"`
	Password   string      `gorm:"size:64"`
	Nickname   string      `gorm:"size:32"`
	Avatar     string      `gorm:"size:256"`
	Role       string      `gorm:"size:256"`
	CreateForm string      `gorm:"size:64"`
	State      string      `gorm:"size:256"`
	Config     *UserConfig `gorm:"foreignKey:UserId;constraint:OnDelete:CASCADE;"`
}

type UserConfig struct {
	ModelBase
	UserId     int64  `gorm:"not null"`
	State      string `gorm:"size:32"`
	Appearance string `gorm:"size:32"`
}
