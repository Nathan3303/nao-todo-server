# 阶段四：边界处理与联调

## 实现内容

- 修复 `UpdateRemindAt` 对空字符串的处理（设置为 NULL）
- 修复 `calculateNextWeekly` 无匹配时返回 nil（而非零值）
- 修复 `GetDueReminders`、`ClearRemindRepeat`、`UpdateRemindAt` 的 nil context 处理（fallback 到 `context.Background()`）
- 确保 GORM 软删除自动过滤 `deleted_at IS NOT NULL`（通过 `ModelBase.DeletedAt`）
- 确认位掩码映射正确：Go `time.Weekday` (Sunday=0) → bit 0=1 与 `WeekdayBitmask` 一致

## 编译验证

- `go build ./...` 通过
- `go vet ./...` 通过
