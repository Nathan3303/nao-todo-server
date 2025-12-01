package controllers

import (
	"naotodoserver/interfaces/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

func BuildResponseData(code int, message string, data any) types.ResponseData {
	return types.ResponseData{
		Status:  http.StatusOK,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func EndWithStatus(ctx *gin.Context, status int, resData types.ResponseData) {
	ctx.JSON(status, resData)
}

func Success(ctx *gin.Context, resData types.ResponseData) {
	EndWithStatus(ctx, http.StatusOK, resData)
}

func Failure(ctx *gin.Context, resData types.ResponseData) {
	EndWithStatus(ctx, http.StatusOK, resData)
}

func DefaultHandler(ctx *gin.Context) {
	Success(ctx, types.ResponseData{
		Code:    200,
		Message: "success",
		Data:    nil,
	})
}
