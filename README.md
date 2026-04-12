# Nao Todo Server

一个功能强大的任务管理系统后端服务，基于 Go 语言开发，提供完整的任务、项目、标签、事件和评论管理功能。

## 项目介绍

Nao Todo Server 是一个现代化的任务管理系统后端，采用 DDD（领域驱动设计）架构，提供 RESTful API 接口，支持用户认证、任务管理、项目协作等功能。

## 技术栈

### 核心技术
- **语言**: Go 1.24.6
- **Web 框架**: Gin 1.11.0
- **ORM**: GORM 1.31.1
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+
- **认证**: JWT (JSON Web Token)
- **日志**: Logrus + 每日滚动日志

### 主要依赖
- `github.com/gin-gonic/gin` - Web 框架
- `gorm.io/gorm` - ORM 框架
- `github.com/go-redis/redis/v8` - Redis 客户端
- `github.com/golang-jwt/jwt/v4` - JWT 认证
- `github.com/bwmarrin/snowflake` - 雪花 ID 生成
- `github.com/robfig/cron/v3` - 定时任务

## 项目架构

项目采用分层架构设计：

```
├── application/     # 应用服务层（用例）
├── domain/          # 领域层（实体、值对象、领域服务、仓库接口）
├── infrastructure/  # 基础设施层（数据库、缓存、外部 API 调用）
├── interfaces/      # 接口层（HTTP 控制器、路由、请求响应类型）
│   ├── controllers/ # 控制器
│   ├── middlewares/ # 中间件
│   ├── routers/     # 路由
│   └── types/       # 请求响应类型
├── cmd/             # 主程序入口
├── conf/            # 配置文件
├── consts/          # 常量定义
├── docs/            # API 文档和架构设计
└── .trae/           # Trae IDE 配置
```

## 快速开始

### 1. 环境准备

- Go 1.24.6+
- MySQL 8.0+
- Redis 6.0+
- Git

### 2. 克隆项目

```bash
git clone <项目地址>
cd nao-todo-server
```

### 3. 配置文件

复制并修改配置文件：

```bash
cp conf/config.yaml.example conf/config.yaml
```

编辑 `conf/config.yaml` 文件，根据您的环境修改数据库、Redis 和服务器配置。

### 4. 数据库初始化

确保 MySQL 服务已启动，创建数据库并配置正确的用户权限。

### 5. 安装依赖

```bash
go mod tidy
```

### 6. 运行项目

#### 方式一：直接运行

```bash
go run cmd/main.go
```

#### 方式二：使用 fresh（热重载）

```bash
# 安装 fresh
go install github.com/pilu/fresh@latest

# 运行项目
fresh
```

### 7. 访问项目

项目启动后，访问以下地址：

- 健康检查：http://localhost:3000/api/ping
- API 文档：查看 `/docs/api.md` 文件
- Postman 集合：`/docs/Nao Todo Server API.postman_collection.json`

## 功能特性

### 用户管理
- 用户注册、登录、退出
- 用户信息查询和更新
- 密码修改
- 头像上传和更新
- 用户激活/禁用

### 任务管理
- 任务的增删改查
- 任务状态管理（待办、进行中、已完成、已放弃等）
- 任务优先级设置
- 任务标签管理
- 任务搜索和过滤
- 分页查询
- 任务恢复（从删除状态恢复）

### 项目管理
- 项目的增删改查
- 项目归档/取消归档
- 项目偏好设置（视图类型、列配置等）

### 标签管理
- 标签的增删改查
- 标签颜色管理
- 标签偏好设置

### 事件管理（检查事项）
- 事件的增删改查
- 事件完成状态管理
- 事件排序

### 评论管理
- 任务评论的增删改查
- 评论置顶
- 评论用户信息展示

### 系统功能
- 请求日志记录
- 客户端信息收集
- IP 地址解析（地理位置）
- 频率限制
- JWT 认证和刷新

## API 文档

详细的 API 文档请查看 [API 文档](./docs/api.md)。

### 文档结构
- 认证接口 (/auth)
- 用户接口 (/user)
- 项目接口 (/projects)
- 标签接口 (/tags)
- 任务接口 (/tasks)
- 事件接口 (/events)
- 评论接口 (/comments)

### 响应格式

#### 成功响应
```json
{
  "code": 10000,
  "message": "操作成功",
  "data": {},
  "pagination": {}
}
```

#### 错误响应
```json
{
  "code": 10001,
  "message": "错误信息",
  "error": "详细错误描述"
}
```

### 错误码范围

| 范围 | 模块 | 说明 |
|------|------|------|
| 10000-19999 | 用户认证 | 登录、注册、权限验证 |
| 20000-29999 | 项目管理 | 项目相关操作 |
| 30000-39999 | 标签管理 | 标签相关操作 |
| 40000-49999 | 任务管理 | 任务相关操作 |
| 50000-59999 | 事件管理 | 事件相关操作 |
| 60000-69999 | 评论管理 | 评论相关操作 |

## 数据库设计

### 主要表结构

#### 用户表 (users)
- id (雪花 ID)
- email (邮箱)
- nickname (昵称)
- avatar (头像 URL)
- role (角色)
- state (状态)
- config (用户配置)
- created_at, updated_at

#### 任务表 (tasks)
- id (雪花 ID)
- parent_task_id (父任务 ID)
- project_id (项目 ID)
- name (任务名称)
- description (任务描述)
- state (任务状态)
- priority (任务优先级)
- start_at, end_at (开始/结束时间)
- tags (标签列表)
- deleted_at, archived_at, star_mark_at, given_up_at

#### 项目表 (projects)
- id (雪花 ID)
- name (项目名称)
- description (项目描述)
- archived_at (归档时间)

#### 标签表 (tags)
- id (雪花 ID)
- name (标签名称)
- description (标签描述)
- color (标签颜色)

#### 事件表 (events)
- id (雪花 ID)
- task_id (所属任务 ID)
- name (事件名称)
- description (事件描述)
- is_done (完成状态)
- sort_id (排序 ID)

#### 评论表 (comments)
- id (雪花 ID)
- task_id (所属任务 ID)
- content (评论内容)
- attachments (附件)
- is_top_up (是否置顶)
- comment_user (评论用户信息)

## 开发规范

### 代码规范
- 遵循 Go 官方代码规范
- 使用 `go fmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查

### 提交规范
- 提交信息应清晰描述修改内容
- 使用语义化提交信息（feat: 新功能, fix: 修复 bug, docs: 文档更新等）

### 分支规范
- `main` - 主分支（生产环境）
- `develop` - 开发分支（测试环境）
- `feature/*` - 功能开发分支
- `hotfix/*` - 紧急修复分支

## 部署

### 生产环境部署

#### 1. 编译二进制文件

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o nao-todo-server cmd/main.go
```

#### 2. 创建服务配置

```systemd
[Unit]
Description=Nao Todo Server
After=network.target

[Service]
Type=simple
User=www-data
Group=www-data
WorkingDirectory=/path/to/project
ExecStart=/path/to/project/nao-todo-server
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

#### 3. 启动服务

```bash
sudo systemctl daemon-reload
sudo systemctl start nao-todo-server
sudo systemctl enable nao-todo-server
```

### Docker 部署

```dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o nao-todo-server ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/nao-todo-server .
COPY --from=builder /app/conf/config.yaml.example conf/config.yaml
COPY --from=builder /app/infrastructure/ip2region/ip2region.xdb infrastructure/ip2region/ip2region.xdb

EXPOSE 3000

CMD ["./nao-todo-server"]
```

## 监控和日志

### 日志配置

日志文件默认保存在 `logs/app.log`，每日自动滚动：

- 单个日志文件最大 100MB
- 保留 30 天的历史日志
- 压缩历史日志文件

### 日志级别

支持以下日志级别（从低到高）：
- debug
- info
- warn
- error
- fatal
- panic

### 健康检查

可以通过以下接口检查服务健康状态：

```
GET /api/ping
```

## 常见问题

### 1. 启动失败

- 检查数据库连接配置是否正确
- 检查 Redis 连接配置
- 确保端口 3000 未被占用

### 2. 认证失败

- 检查 JWT 密钥配置
- 确保 Token 格式正确
- 检查 Token 是否过期

### 3. 数据库迁移

项目使用 GORM 的自动迁移功能，第一次启动时会自动创建表。如果需要手动迁移：

```bash
go run cmd/migrate.go
```

### 4. IP 解析失败

确保 `infrastructure/ip2region/ip2region.xdb` 文件存在且可读取。

## 开发路线图

- [ ] 添加任务提醒功能
- [ ] 支持任务协作和共享
- [ ] 添加任务附件管理
- [ ] 实现任务统计和图表
- [ ] 支持多语言
- [ ] 添加 API 文档自动生成 (Swagger)
- [ ] 实现服务监控和报警
- [ ] 优化查询性能

## 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交修改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](./LICENSE) 文件。

MIT License 是一种宽松的开源许可证，允许您：
- 免费使用、复制、修改、合并、出版、分发、再授权和销售本软件
- 在任何个人或商业项目中使用本软件
- 无需向原作者支付任何费用

唯一的条件是，您必须在所有副本或重要部分的软件中保留原始版权声明和许可证声明。

有关 MIT 许可证的完整内容，请查看 LICENSE 文件。

## 联系方式

如有问题或建议，欢迎通过以下方式联系：

- 邮箱: [your-email@example.com]
- GitHub: [项目地址]
- 文档: `/docs/` 目录下的架构设计和 API 文档

---

**注**: 这是一个持续开发中的项目，功能和文档会定期更新。
