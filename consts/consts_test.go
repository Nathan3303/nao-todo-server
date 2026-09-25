package consts

import (
	"testing"
	"time"
)

// TestSnowflakeEpochMS 锁定雪花 Epoch 跨端契约值：
// 1. 与 2023-03-01T00:00:00Z 计算值一致
// 2. 锁定字面量 1677628800000，防止未来误改导致双端 ID 时间语义分裂
func TestSnowflakeEpochMS(t *testing.T) {
	computed := time.Date(2023, 3, 1, 0, 0, 0, 0, time.UTC).Unix() * 1000
	if SnowflakeEpochMS != computed {
		t.Fatalf("SnowflakeEpochMS = %d, want computed %d", SnowflakeEpochMS, computed)
	}
	if SnowflakeEpochMS != 1677628800000 {
		t.Fatalf("SnowflakeEpochMS = %d, want locked literal 1677628800000", SnowflakeEpochMS)
	}
}
