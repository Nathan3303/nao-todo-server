# 阶段二：Hub + SSE

## 实现内容

- 新建 SSE Hub 单例 (`infrastructure/sse/hub.go`)：管理 userId→channel 映射，提供 Publish/Subscribe/Unsubscribe
- 新建 SSE controller (`interfaces/controllers/sse.go`)：`ReminderStream` handler，从 Hub 订阅当前用户事件并通过 SSE 推送
- 新建 SSE router (`interfaces/routers/sseRouter.go`)：注册 `GET /api/sse/reminders`，JWT 鉴权
- 在主路由初始化中注册 SSE 路由组

## 受影响文件

- `infrastructure/sse/hub.go` — 新建
- `interfaces/controllers/sse.go` — 新建
- `interfaces/routers/sseRouter.go` — 新建
- `interfaces/routers/routers.go` — 添加 UseSSERouter 调用
