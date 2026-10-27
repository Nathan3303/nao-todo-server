package identity

import (
	"sync"
	"time"
)

// localSweepInterval 本地兜底表过期清理周期：
// 键由客户端 IP / 账号哈希产生，必须周期清理以免离群请求把内存撑大。
const localSweepInterval = 30 * time.Second

// localCounter 固定窗口计数项
type localCounter struct {
	count    int64
	expireAt time.Time
}

// LocalFallback 进程内兜底计数器（SEC-10 / SEC-04）。
// Redis 不可用时由限流与账号失败节流共用，仅本实例生效
// （多实例计数不同步，属已声明的局限）。
type LocalFallback struct {
	mu        sync.Mutex
	counters  map[string]*localCounter
	lastSweep time.Time
}

// NewLocalFallback 创建进程内兜底计数器
func NewLocalFallback() *LocalFallback {
	return &LocalFallback{
		counters:  make(map[string]*localCounter),
		lastSweep: time.Now(),
	}
}

// Incr 计数 +1 并返回当前窗口计数；键不存在或已过期时从 1 重新开始
// @param key 计数键
// @param ttl 窗口长度（首次写入时确定，到期自动失效）
// @return 当前窗口计数
func (l *LocalFallback) Incr(key string, ttl time.Duration) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	l.sweepLocked(now)
	counter, ok := l.counters[key]
	if !ok || now.After(counter.expireAt) {
		counter = &localCounter{expireAt: now.Add(ttl)}
		l.counters[key] = counter
	}
	counter.count++
	return counter.count
}

// Count 返回键当前窗口的计数（不存在或已过期视为 0）
// @param key 计数键
// @return 当前窗口计数
func (l *LocalFallback) Count(key string) int64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	counter, ok := l.counters[key]
	if !ok || time.Now().After(counter.expireAt) {
		return 0
	}
	return counter.count
}

// Reset 删除键（登录成功后清零）
// @param key 计数键
func (l *LocalFallback) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.counters, key)
}

// sweepLocked 按周期清理过期键（调用方需持有锁）
// @param now 当前时间
func (l *LocalFallback) sweepLocked(now time.Time) {
	if now.Sub(l.lastSweep) < localSweepInterval {
		return
	}
	for key, counter := range l.counters {
		if now.After(counter.expireAt) {
			delete(l.counters, key)
		}
	}
	l.lastSweep = now
}
