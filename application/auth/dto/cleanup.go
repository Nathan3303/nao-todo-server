package dto

import "time"

// CleanupSessionsOptions 过期会话清理入参
// 时间上界由调用方（cron）按配置计算，应用层不读配置，便于单测注入固定时间
type CleanupSessionsOptions struct {
	// ExpiredBefore 过期会话回收上界：仅 expired_at < ExpiredBefore 的行会被回收
	ExpiredBefore time.Time
	// RevokedBefore 已登出会话回收上界：仅 deleted_at < RevokedBefore 的行会被回收
	RevokedBefore time.Time
	// BatchSize 单批删除行数上限
	BatchSize int
	// MaxBatches 单次执行最多批数（过期 + 已登出共享该预算）
	MaxBatches int
	// BatchPause 批间暂停时长
	BatchPause time.Duration
}
