package auth

import (
	"strconv"
	"time"

	"naotodoserver/application/auth/dto"
	"naotodoserver/domain/identity/entities"
	"naotodoserver/domain/identity/valueobjects"
)

// SignUpInputToCreateUserValueObject 将 SignUpInput 转换为 CreateUserValueObject
// @param signUpInput 注册入参
// @return CreateUserValueObject 创建用户值对象
// @error 错误信息
func SignUpInputToCreateUserValueObject(
	signUpInput *dto.SignUpInput,
) (*valueobjects.CreateUser, error) {
	return valueobjects.NewCreateUser(
		signUpInput.Email,
		signUpInput.Password,
		signUpInput.Email,
		signUpInput.Nickname,
	)
}

// SessionEntity2Item 会话实体转换为会话列表项
// @param e 会话实体
// @param currentToken 当前请求的会话令牌（用于标记当前会话）
// @return 会话列表项
func SessionEntity2Item(e *entities.UserSession, currentToken string) *dto.SessionItem {
	return &dto.SessionItem{
		Id:         strconv.FormatInt(e.Id, 10),
		DeviceId:   e.DeviceId,
		DeviceType: e.DeviceType,
		IP4:        e.IP4,
		Region:     e.Region,
		CreatedAt:  e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  e.UpdatedAt.Format(time.RFC3339),
		Current:    e.Token == currentToken,
	}
}
