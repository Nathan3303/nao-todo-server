# Nao Todo Server

一个基于 Go 与 DDD 架构的待办事项 / 清单 / 标签 / 番茄钟后端服务，提供 RESTful API 与 SSE 实时推送。

## 特性

- 完整的 DDD 四层架构：`interfaces` / `application` / `domain` / `infrastructure`
- 用户认证：注册、登录、JWT 鉴权、刷新、注销
- 待办任务、清单（项目）、标签、检查事项、评论的 CRUD
- 番茄钟（自定义方案 + 实际记录）
- 任务提醒：定时扫描 + SSE 实时推送
- 文件上传：本地头像存储，`/static/uploads` 静态映射
- 安全响应头、Gzip 压缩、CORS、IP 限流
- MySQL + Redis 持久化与缓存
- GORM AutoMigrate，自动建表
- 滚动日志、单文件大小限制、自动压缩

## 技术栈

| 类别 | 选型 |
| ---- | ---- |
| 语言 | Go 1.25 |
| Web 框架 | [Gin](https://github.com/gin-gonic/gin) |
| ORM | [GORM](https://gorm.io) v1.31 |
| 数据库 | MySQL 8.0+ |
| 缓存 | Redis 6.0+ |
| 认证 | JWT (golang-jwt/jwt v4) |
| 配置 | Viper |
| 日志 | Logrus + file-rotatelogs |
| 定时任务 | robfig/cron v3 |
| ID 生成 | bwmarrin/snowflake |
| IP 解析 | lionsoul2014/ip2region |
| 实时推送 | Server-Sent Events (Gin SSE) |

## 架构

```
nao-todo-server/
├── cmd/                # 入口（main.go）
├── conf/               # 配置文件与配置加载
├── application/        # 应用服务层（用例编排）
│   ├── auth/           #   认证用例
│   ├── user/           #   用户用例
│   ├── project/        #   清单用例
│   ├── task/           #   任务 / 检查事项 / 评论用例
│   ├── tag/            #   标签用例
│   └── pomodoro/       #   番茄钟用例
├── domain/             # 领域层（核心业务）
│   ├── identity/       #   用户与认证领域
│   ├── project/        #   清单领域
│   ├── task/           #   任务领域
│   ├── tag/            #   标签领域
│   ├── pomodoro/       #   番茄钟领域
│   ├── errors/         #   领域错误定义
│   └── types/          #   通用类型
├── infrastructure/     # 基础设施层
│   ├── persistence/    #   MySQL / Redis 实现
│   ├── auth/           #   JWT 实现
│   ├── cron/           #   定时任务
│   ├── ip2region/      #   IP 地理解析
│   ├── logging/        #   日志
│   ├── sse/            #   事件总线
│   └── storage/        #   头像本地存储
├── interfaces/         # 接口层
│   ├── controllers/    #   HTTP 控制器
│   ├── middlewares/    #   中间件（JWT、限流、日志…）
│   ├── routers/        #   路由注册
│   └── types/          #   请求 / 响应 DTO
├── docs/               # 架构文档与开发计划
└── .trae/              # Trae IDE 规则与计划
```

依赖方向：`interfaces` → `application` → `domain` ← `infrastructure`。

## 快速开始

### 环境要求

- Go 1.25+
- MySQL 8.0+
- Redis 6.0+

### 1. 获取代码

```bash
git clone https://github.com/<your-org>/nao-todo-server.git
cd nao-todo-server
```

### 2. 配置

复制示例配置并按需修改：

```bash
cp .example/.env.example .env
cp .example/app.config.yaml.example conf/config.yaml
```

`conf/config.yaml` 关键字段：

| 节点 | 说明 |
| ---- | ---- |
| `server.port` | HTTP(S) 端口（默认 3302） |
| `server.certFile` / `keyFile` | 同时配置时启用 HTTPS |
| `mysql.*` | MySQL 连接信息 |
| `redis.*` | Redis 连接信息 |
| `log.*` | 日志级别、文件路径、滚动策略 |
| `uploads.*` | 头像 / 附件存储目录与大小限制 |

### 3. 启动依赖

最简单的方式是使用项目自带的 Docker Compose：

```bash
docker compose -f docker-compose-dev.yml up -d
```

或参考 `.example/` 下 `my.cnf.example` / `redis.conf.example` 在本机直接启动。

### 4. 启动服务

```bash
go mod tidy
go run cmd/main.go
```

或使用 `fresh` 实现热重载：

```bash
go install github.com/pilu/fresh@latest
fresh
```

### 5. 健康检查

```bash
curl http://localhost:3302/api/ping
```

## 配置项说明

```yaml
server:
  ip: ""              # 留空表示监听 0.0.0.0
  port: 3302          # HTTP/HTTPS 端口
  version: 1.0
  jwtSecret: <密钥>    # JWT 签名密钥
  goMaxProc: 2        # Go 运行时最大 P 数
  certFile: ""        # TLS 证书；与 keyFile 同时存在时启用 HTTPS
  keyFile: ""
  debug: true         # false 时切换为 Release 模式
```

> 当 `certFile` 与 `keyFile` 都配置时，服务以 HTTPS 启动，并启用 TLS 1.2 + 现代加密套件。

## API 概览

所有接口均挂在 `/api` 前缀下，统一使用 `ClientInfo` + `RequestLogger` 中间件；除 `ping` 与认证接口外，其余都需经过 `JWTValidator` 与 `RateLimiter` 中间件。

| 模块 | 路径前缀 | 说明 |
| ---- | -------- | ---- |
| 健康检查 | `GET /api/ping` | 无需鉴权 |
| 认证 | `/api/auth/*` | 注册、登录、刷新、注销 |
| 用户 | `/api/user/*` | 用户信息、密码、头像 |
| 清单 | `/api/projects/*` | 清单 CRUD、归档、偏好 |
| 标签 | `/api/tags/*` | 标签 CRUD、偏好 |
| 任务 | `/api/tasks/*` | 任务 CRUD、状态、优先级、恢复 |
| 检查事项 | `/api/events/*` | 子事项的 CRUD 与排序 |
| 评论 | `/api/comments/*` | 评论 CRUD、置顶 |
| 番茄钟 | `/api/pomodoros/*` `/api/pomodoro-records/*` | 方案与实际记录 |
| 实时推送 | `GET /api/sse/reminders` | 任务到期提醒 |

### 响应格式

成功：

```json
{
  "code": 10000,
  "message": "操作成功",
  "data": {},
  "pagination": { "total": 0, "page": 1, "limit": 20, "maxPage": 0 }
}
```

失败：

```json
{
  "code": 10001,
  "error": "详细错误描述"
}
```

### 错误码分段

| 区间 | 模块 |
| ---- | ---- |
| 10000–19999 | 用户与认证 |
| 20000–29999 | 清单 / 项目 |
| 30000–39999 | 标签 |
| 40000–49999 | 任务 |
| 50000–59999 | 检查事项 |
| 60000–69999 | 评论 |
| 70000–79999 | 番茄钟 |

## 部署

### 直接部署

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o nao-todo-server cmd/main.go
```

推荐使用 `systemd` 托管，可参考 `docs/plans/server-deploy.md`。

### Docker

```bash
docker build -t nao-todo-server:latest .
docker run -d \
  --name nao-todo-server \
  -p 3302:3302 \
  -v $(pwd)/conf:/app/conf \
  -v $(pwd)/infrastructure/ip2region:/app/infrastructure/ip2region \
  -v $(pwd)/uploads:/app/uploads \
  nao-todo-server:latest
```

## 日志与监控

- 日志文件：`logs/app.log`
- 单文件最大 100MB，保留 30 天，压缩归档
- 支持级别：`debug` / `info` / `warn` / `error` / `fatal` / `panic`
- 健康检查：`GET /api/ping`

## 常见问题

- **启动失败 / 端口占用**：检查 `conf/config.yaml` 的 `server.port`、MySQL / Redis 是否可达。
- **JWT 鉴权失败**：确认 `jwtSecret` 与签发端一致、Token 未过期、请求头携带 `Authorization: Bearer <token>`。
- **数据库未建表**：项目使用 GORM AutoMigrate，首次启动会自动建表；如需手动控制请自行管理 migration。
- **IP 解析失败**：确认 `infrastructure/ip2region/ip2region.xdb` 存在且可读。
- **SSE 不工作**：当前在 Nginx 反代时需关闭缓冲，可参考 `.example/nginx.sub.conf.example`。

## 开发规范

- 遵循 Go 官方代码风格，`gofmt` / `goimports` 格式化
- 提交前运行 `golangci-lint run`
- 分支策略：`main`（生产）/ `develop`（集成）/ `feature/*` / `hotfix/*`
- 领域代码不依赖任何框架（仅可引用 `domain` 内部包）

## 路线图

- [ ] Swagger / OpenAPI 文档自动生成
- [ ] 任务附件管理
- [ ] 任务统计与图表数据接口
- [ ] 多语言支持
- [ ] 服务指标暴露（Prometheus）

## 许可证

本项目基于 [MIT](./LICENSE) 许可证发布。

Copyright (c) 2026 Nathan Lee
