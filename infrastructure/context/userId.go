package context

import "context"

type ktype string

var contextKey ktype = "userId"

func GetUserId(ctx context.Context) int64 {
	v, ok := ctx.Value(contextKey).(int64)
	if !ok {
		return 0
	}
	return v
}

func SetUserId(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, contextKey, userId)
}
