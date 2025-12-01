package context

import "context"

type key uint

var contextKey key

func GetUserId(ctx context.Context) int64 {
	userId, ok := ctx.Value(contextKey).(int64)
	if !ok {
		return 0
	}
	return userId
}

func SetUserId(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, contextKey, userId)
}
