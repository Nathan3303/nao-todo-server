package context

import "context"

type requestIdKeyType string

var requestIdKey requestIdKeyType = "requestId"

// GetRequestId 读取当前请求的关联标识（未设置时为空串）
// @param ctx 上下文
// @return 关联标识
func GetRequestId(ctx context.Context) string {
	v, ok := ctx.Value(requestIdKey).(string)
	if !ok {
		return ""
	}
	return v
}

// SetRequestId 写入当前请求的关联标识
// @param ctx 上下文
// @param requestId 关联标识（服务端生成或来自已校验的 X-Request-Id）
// @return 新上下文
func SetRequestId(ctx context.Context, requestId string) context.Context {
	return context.WithValue(ctx, requestIdKey, requestId)
}
