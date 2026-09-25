# 阶段三：扫描与推送

## 实现内容

- 实现 `CalculateNextRemindAt` 函数（domain service）：根据 repeat 类型计算下一次提醒时间
  - daily: +1 天
  - weekly: 扫描 1~7 天，匹配 weekdays 位掩码
  - monthly: +1 月
  - 超出 end_at 或无匹配 weekday 返回 nil
- 新建 ReminderJob cron 任务 (`infrastructure/cron/reminder.go`)：
  - 每分钟查询 `remind_at <= NOW()` 的未删除任务
  - 通过 Hub.Publish 推送 SSE 事件
  - 重复提醒：计算 next → 更新 remind_at 或清除规则
  - 单次提醒：清除 remind_at
- 仓库接口新增 `GetDueReminders`、`ClearRemindRepeat`、`UpdateRemindAt` 方法
- `LoadCron()` 注册提醒扫描任务（每分钟执行）

## 受影响文件

- `domain/task/service/serviceImpl.go` — 新增 CalculateNextRemindAt
- `domain/task/repositories/task.go` — 新增 3 个接口方法
- `infrastructure/persistence/task/repoImpl.go` — 实现 3 个新方法
- `infrastructure/cron/reminder.go` — 新建
- `infrastructure/initialize.go` — 注册 cron job
