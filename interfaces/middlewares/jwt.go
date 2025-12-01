package middlewares

import (
	"naotodoserver/application/user"
	"naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"
	"strings"

	"github.com/gin-gonic/gin"
)

func getJwtString(ctx *gin.Context) string {
	jwtRaw := ctx.GetHeader("Authorization")
	jwtString := strings.Split(jwtRaw, "Bearer ")
	jwtString = append(jwtString, "")
	return jwtString[1]
}

func JWTValidator(ctx *gin.Context) {
	// @step 1. 获取 JWT
	jwtString := getJwtString(ctx)

	// @step 2. 验证 JWT 并处理失败结果
	err, jwtPayload := user.UserService.UserDomain.ValidateJWT(ctx, jwtString)
	if err != nil {
		controllers.Failure(ctx, types.ResponseData{
			Code:    10042,
			Message: "用户凭证无效",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// @step 3. 验证 Session 会话记录,并处理需要重新登录的结果
	isSessionValid, _ := user.UserService.UserDomain.ValidateSession(ctx, jwtString)
	if !isSessionValid {
		controllers.Failure(ctx, types.ResponseData{
			Code:    10043,
			Message: "用户凭证已过期,请重新登录",
			Data:    nil,
		})
		ctx.Abort()
		return
	}

	// @step 4. 写入用户信息到上下文
	ctx.Request = ctx.Request.WithContext(
		context.SetUserId(ctx.Request.Context(), jwtPayload.Id),
	)

	// @step 5. 检测通过，继续处理请求
	ctx.Next()
}
