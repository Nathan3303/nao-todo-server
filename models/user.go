package models

type User struct {
	Model

	Account    string `gorm:"size:32 unique" json:"account"`
	Email      string `gorm:"size:64 unique" json:"email"`
	Password   string `gorm:"size:64" json:"password"`
	Nickname   string `gorm:"size:32" json:"nickname"`
	Avatar     string `gorm:"size:256" json:"avatar"`
	Role       string `gorm:"size:256" json:"role"`
	CreateForm string `gorm:"size:32" json:"createForm"`
	State      string `gorm:"size:256" json:"state"`
}

type UserConfig struct {
	Model

	UserId int64 `gorm:"not null" json:"userId"`

	State      string `gorm:"size:16" json:"state"`
	Appearance string `gorm:"size:16" json:"appearance"`
}

type UserResponse struct {
	ModelResponse

	Account  string `json:"account"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}
