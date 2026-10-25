package cron

import (
	"context"
	"fmt"
	"time"

	authApp "naotodoserver/application/auth"
	"naotodoserver/application/auth/dto"
	"naotodoserver/conf"
)

// sessionCleanupBatchPause 批间暂停时长（评审 T4：500 行/批、批间 50ms）
const sessionCleanupBatchPause = 50 * time.Millisecond

// CleanExpiredSessionsJob 过期会话清理任务
// 只做壳：计算时间上界后委派应用层按批删除；无 DDL、不改写 expired_at。
type CleanExpiredSessionsJob struct {
	AuthApp authApp.AuthApp
	// Cfg 清理配置（宽限期 / 批量 / 批数）
	Cfg conf.SessionCleanup
	// Now 当前时间来源（单测可注入固定时钟）
	Now func() time.Time
}

// NewCleanExpiredSessionsJob 创建过期会话清理任务
func NewCleanExpiredSessionsJob(
	authApp authApp.AuthApp,
	cfg conf.SessionCleanup,
) *CleanExpiredSessionsJob {
	return &CleanExpiredSessionsJob{AuthApp: authApp, Cfg: cfg, Now: time.Now}
}

// Run 执行清理
// @note 时间上界在每次执行时计算（cron 长期驻留，不能用注册时的绝对时间）
func (ces *CleanExpiredSessionsJob) Run() {
	now := ces.Now()
	opts := dto.CleanupSessionsOptions{
		ExpiredBefore: now.AddDate(0, 0, -ces.Cfg.RetentionDays),
		RevokedBefore: now.Add(-time.Duration(ces.Cfg.RevokedRetentionHours) * time.Hour),
		BatchSize:     ces.Cfg.BatchSize,
		MaxBatches:    ces.Cfg.MaxBatches,
		BatchPause:    sessionCleanupBatchPause,
	}
	removedExpired, removedRevoked, err := ces.AuthApp.CleanupExpiredSessions(
		context.TODO(),
		opts,
	)
	if err != nil {
		fmt.Println("清理过期会话失败：" + err.Error())
		return
	}
	// 部署验证对账用：无 token 等敏感字段，只报行数
	fmt.Printf(
		"session_cleanup removed_expired=%d removed_revoked=%d\n",
		removedExpired,
		removedRevoked,
	)
}
