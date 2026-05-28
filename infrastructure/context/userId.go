package context

import "context"

type ktype string

var (
	userIdKey ktype = "userId"
	tokenKey  ktype = "token"
)

func GetUserId(ctx context.Context) int64 {
	v, ok := ctx.Value(userIdKey).(int64)
	if !ok {
		return 0
	}
	return v
}

func SetUserId(ctx context.Context, userId int64) context.Context {
	return context.WithValue(ctx, userIdKey, userId)
}

func GetToken(ctx context.Context) string {
	v, ok := ctx.Value(tokenKey).(string)
	if !ok {
		return ""
	}
	return v
}

func SetToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey, token)
}
