package controllers

import (
	"io"
	"naotodoserver/infrastructure/context"
	"naotodoserver/infrastructure/sse"
	"naotodoserver/interfaces/types"

	"github.com/gin-gonic/gin"
)

// ReminderStream 提醒事件 SSE 流
// @code 5000x
func ReminderStream(ctx *gin.Context) {
	// 1. 获取用户 ID
	userId := context.GetUserId(ctx.Request.Context())
	if userId <= 0 {
		Failure(ctx, types.ResponseData{
			Code:    50001,
			Message: "用户 ID 无效",
		})
		return
	}
	// 2. 设置 SSE 响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("X-Accel-Buffering", "no")
	// 3. 订阅 Hub
	hub := sse.GetHub()
	ch := hub.Subscribe(userId)
	defer hub.Unsubscribe(userId, ch)
	// 4. 流式写入事件
	ctx.Stream(func(w io.Writer) bool {
		select {
		case event := <-ch:
			ctx.SSEvent("reminder", string(event))
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
}
