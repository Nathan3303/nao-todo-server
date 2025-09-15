package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Status  int    `json:"-"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func BuildResponseData(code int, message string, data any) ResponseData {
	return ResponseData{
		Status:  http.StatusOK,
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func EndWithStatus(ctx *gin.Context, status int, resData ResponseData) {
	ctx.JSON(status, resData)
}

func Success(ctx *gin.Context, resData ResponseData) {
	EndWithStatus(ctx, http.StatusOK, resData)
}

func Failure(ctx *gin.Context, resData ResponseData) {
	EndWithStatus(ctx, http.StatusOK, resData)
}

func DefaultHandler(ctx *gin.Context) {
	Success(ctx, ResponseData{
		Code:    200,
		Message: "success",
		Data:    nil,
	})
}
