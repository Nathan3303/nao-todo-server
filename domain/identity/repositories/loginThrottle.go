package repositories

import (
	"context"
	"time"
)

// LoginThrottle 账号维度登录失败节流（SEC-04）。
// 键由调用方用 valueobjects.AccountKey 生成，实现侧不得落邮箱明文。
// 实现契约：Redis 不可用时**不阻断登录**（fail-open），退化为进程内计数，
// 因此各方法不返回错误——降级由实现自行记录日志。
type LoginThrottle interface {
	// RecordFailure 记录一次登录失败，返回窗口内累计失败次数
	RecordFailure(ctx context.Context, accountKey string, window time.Duration) int64
	// Block 将账号置为暂停登录状态，duration 到期后自动解除
	Block(ctx context.Context, accountKey string, duration time.Duration)
	// IsBlocked 轻量查询账号是否处于暂停登录状态（单次查询，不查库）
	IsBlocked(ctx context.Context, accountKey string) bool
	// Reset 清除账号的失败计数与暂停状态（登录成功后调用）
	Reset(ctx context.Context, accountKey string)
}
