package apis

import (
	"naotodoserver/models"
	"strconv"
)

func ToCommentResponseList(comments []models.Comment) []models.CommentResponse {
	responseList := make([]models.CommentResponse, len(comments))
	for i, item := range comments {
		responseList[i] = ToCommentResponse(&item)
	}
	return responseList
}

func ToCommentResponse(comment *models.Comment) models.CommentResponse {
	var commentResponse models.CommentResponse
	commentResponse.ID = strconv.FormatInt(comment.ID, 10)
	commentResponse.CreatedAt = comment.CreatedAt
	commentResponse.UpdatedAt = comment.UpdatedAt
	commentResponse.DeletedAt = comment.DeletedAt

	commentResponse.UserId = strconv.FormatInt(comment.UserId, 10)
	commentResponse.TodoId = strconv.FormatInt(comment.TodoId, 10)

	commentResponse.Content = comment.Content
	commentResponse.Attachments = comment.Attachments
	commentResponse.IsTopUp = comment.IsTopUp

	if comment.CommentUser != nil {
		commentResponse.CommentUser = &models.CommentUserResponse{
			Nickname: comment.CommentUser.Nickname,
			Avatar:   comment.CommentUser.Avatar,
		}
	}

	return commentResponse
}
