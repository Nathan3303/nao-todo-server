# Checklist

- [x] 缓存辅助组件基于 `dbs.RdsCli`，提供带 TTL 的 Get/Set/Del 与用户维度 key 构造
- [x] Redis 出错时读取按未命中处理、写/删静默失败，调用方回源 MySQL 且请求正常完成
- [x] 会话校验：`IsSessionValid` 结果缓存命中，Create/UpdateToken/Delete 后失效
- [x] 用户资料/配置：`FindById`/`GetConfig` 缓存命中，相关写操作后失效
- [x] 项目列表：`GetByUserId` 缓存命中，Create/Update/Delete/Restore/Archive/Unarchive/BatchUpdate 后失效
- [x] 标签列表：`Get` 缓存命中，Create/Update/Delete/BatchUpdate 后失效
- [x] `initialize.go` 已将 `dbs.RdsCli` 注入相关仓储，构造函数签名同步更新
- [x] 所有缓存 key 以 userId 维度隔离，不同用户数据不串
- [x] 代码遵循项目现有风格（中文注释、go-redis/v8、错误处理方式一致）
- [x] `go build ./...` 通过
