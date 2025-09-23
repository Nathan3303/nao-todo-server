package apis

import (
	"naotodoserver/models"
	"strconv"
)

func ToUserResponse(user *models.User) models.UserResponse {
	var userResponse models.UserResponse

	userResponse.ID = strconv.FormatInt(user.ID, 10)
	userResponse.CreatedAt = user.CreatedAt
	userResponse.UpdatedAt = user.UpdatedAt
	userResponse.DeletedAt = user.DeletedAt

	userResponse.Account = user.Account
	userResponse.Email = user.Email
	userResponse.Nickname = user.Nickname
	userResponse.Avatar = user.Avatar
	userResponse.Role = user.Role

	return userResponse
}
