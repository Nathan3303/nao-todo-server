package identity

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"naotodoserver/infrastructure/logging"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// newDownRedisClient 返回一个必然连接失败的 Redis 客户端（端口 1，禁用重试），
// 用于验证限流 / 账号节流的降级路径。
func newDownRedisClient(t *testing.T) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  20 * time.Millisecond,
		ReadTimeout:  20 * time.Millisecond,
		WriteTimeout: 20 * time.Millisecond,
		MaxRetries:   -1,
	})
	t.Cleanup(func() {
		_ = client.Close()
	})
	return client
}

// captureLogs 把全局日志器替换为写入缓冲区的 JSON 日志器，
// 返回缓冲区并提供测试结束时的还原。
func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	previous := logging.Logger
	logger := logrus.New()
	logger.SetOutput(buf)
	logger.SetFormatter(&logrus.JSONFormatter{TimestampFormat: "2006-01-02 15:04:05"})
	logger.SetLevel(logrus.DebugLevel)
	logging.Logger = logger
	t.Cleanup(func() {
		logging.Logger = previous
	})
	return buf
}

// logEntries 解析缓冲区中的 JSON 日志行
func logEntries(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	entries := make([]map[string]any, 0, 4)
	for _, line := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		entry := map[string]any{}
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("日志不是结构化 JSON: %v（line=%s）", err, line)
		}
		entries = append(entries, entry)
	}
	return entries
}

// findLogEntry 按 event 字段取第一条匹配的日志
func findLogEntry(entries []map[string]any, event string) map[string]any {
	for _, entry := range entries {
		if entry["event"] == event {
			return entry
		}
	}
	return nil
}
