package middlewares

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// BodyLimit 全局请求体上限中间件（SEC-01 / T561）。
//
// 机制：有界预读 + 还原 Body。`http.MaxBytesReader` 保证最多只驻留 limit+1 字节，
// 不会像裸 `io.ReadAll` 那样把 10 MB 级 body 读进内存（单请求 OOM 的根因）。
// 超限直接 413 并中断链路（不静默截断）；未超限则把完整 body 还原给下游
// ShouldBind / io.ReadAll，既有控制器零改动。
//
// @param defaultBytes 默认上限（字节）；<=0 表示不限制
// @param overrides 按 URL.Path 精确匹配的单独上限（如头像上传放宽到 5 MB）
func BodyLimit(defaultBytes int64, overrides map[string]int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		limit := defaultBytes
		if v, ok := overrides[ctx.Request.URL.Path]; ok && v > 0 {
			limit = v
		}
		if limit <= 0 || ctx.Request.Body == nil {
			ctx.Next()
			return
		}
		// Content-Length 已声明超限：不读 body，直接拒绝（省掉一次 2 MB 级传输）
		if ctx.Request.ContentLength > limit {
			abortTooLarge(ctx, limit)
			return
		}
		body, err := io.ReadAll(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, limit))
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				abortTooLarge(ctx, limit)
				return
			}
			controllers.FailureByHttpStatus(ctx, http.StatusBadRequest, types.ResponseData{
				Code:    http.StatusBadRequest,
				Message: "请求体读取失败",
			})
			ctx.Abort()
			return
		}
		ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
		ctx.Next()
	}
}

// abortTooLarge 返回 413（不静默截断），并中断后续处理器。
// @param ctx gin 上下文
// @param limit 命中的上限（仅用于提示，便于客户端/运维定位）
func abortTooLarge(ctx *gin.Context, limit int64) {
	controllers.FailureByHttpStatus(ctx, http.StatusRequestEntityTooLarge, types.ResponseData{
		Code:    http.StatusRequestEntityTooLarge,
		Message: "请求体超出上限",
		Error:   fmt.Sprintf("请求体最大 %d 字节", limit),
	})
	ctx.Abort()
}
