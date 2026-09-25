package entities

// UserRole 用户角色
type UserRole uint8

const (
	UserRoleNormal UserRole = 0
	UserRoleAdmin  UserRole = 1
)

// UserState 用户状态
type UserState uint8

const (
	UserStateActive      UserState = 0
	UserStateDeactivated UserState = 1
)
