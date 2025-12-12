package repositories

import "context"

type RateLimit interface {
	Get(ctx context.Context, key string) int8
	Incr(ctx context.Context, key string) error
}
