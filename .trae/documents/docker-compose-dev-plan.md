# Docker Compose 开发环境配置计划

## 需求分析

根据现有的 `docker-compose.yml`，生成一个可在本地 Windows 环境运行的 `docker-compose-dev.yaml`，并将 `.example` 目录中的示例配置文件输出到 `~/Docker/nao-todo-server/` 对应目录。

**关键变更点：**
- 将所有卷路径从生产环境的 `/opt/nao-todo-server/xxx` 改为开发环境的 `~/Docker/nao-todo-server/xxx`
- 保持服务配置、网络配置基本不变
- Windows 系统下 Docker 使用 `~/` 会映射到用户目录
- 将 `.example` 目录中的配置文件复制到开发环境目录，并去掉 `.example` 后缀

## 文件分析

### 当前 docker-compose.yml 结构

**服务列表：**
1. **mysql** - MySQL 数据库服务
2. **redis** - Redis 缓存服务
3. **app** - 应用程序服务

**卷配置（需要修改路径）：**
- mysql-data: `/opt/nao-todo-server/mysql/data` → `~/Docker/nao-todo-server/mysql/data`
- redis-data: `/opt/nao-todo-server/redis/data` → `~/Docker/nao-todo-server/redis/data`
- app-uploads-data: `/opt/nao-todo-server/app/uploads` → `~/Docker/nao-todo-server/app/uploads`
- app-logs-data: `/opt/nao-todo-server/app/logs` → `~/Docker/nao-todo-server/app/logs`
- app-ssl-certs: `/opt/nao-todo-server/app/ssl` → `~/Docker/nao-todo-server/app/ssl`

**配置文件挂载（需要修改路径）：**
- MySQL 配置: `/opt/nao-todo-server/mysql/conf/my.cnf` → `~/Docker/nao-todo-server/mysql/conf/my.cnf`
- MySQL 客户端配置: `/opt/nao-todo-server/mysql/conf/client.cnf` → `~/Docker/nao-todo-server/mysql/conf/client.cnf`
- Redis 配置: `/opt/nao-todo-server/redis/conf/redis.conf` → `~/Docker/nao-todo-server/redis/conf/redis.conf`

### .example 目录文件清单

| 源文件 | 目标文件 |
|--------|----------|
| `.example/my.cnf.example` | `~/Docker/nao-todo-server/mysql/conf/my.cnf` |
| `.example/client.cnf.example` | `~/Docker/nao-todo-server/mysql/conf/client.cnf` |
| `.example/redis.conf.example` | `~/Docker/nao-todo-server/redis/conf/redis.conf` |
| `.example/.env.example` | `~/Docker/nao-todo-server/.env` |

## 实施步骤

### 步骤 1：创建开发环境目录结构

创建以下目录：
- `~/Docker/nao-todo-server/mysql/conf/`
- `~/Docker/nao-todo-server/mysql/data/`
- `~/Docker/nao-todo-server/redis/conf/`
- `~/Docker/nao-todo-server/redis/data/`
- `~/Docker/nao-todo-server/app/uploads/`
- `~/Docker/nao-todo-server/app/logs/`
- `~/Docker/nao-todo-server/app/ssl/`

### 步骤 2：复制配置文件

将 `.example` 目录中的配置文件复制到对应目录，并去掉 `.example` 后缀：

1. `.example/my.cnf.example` → `~/Docker/nao-todo-server/mysql/conf/my.cnf`
2. `.example/client.cnf.example` → `~/Docker/nao-todo-server/mysql/conf/client.cnf`
3. `.example/redis.conf.example` → `~/Docker/nao-todo-server/redis/conf/redis.conf`
4. `.example/.env.example` → `~/Docker/nao-todo-server/.env`

### 步骤 3：创建 docker-compose-dev.yaml

基于现有配置，修改所有卷路径为 `~/Docker/nao-todo-server/xxx` 格式。

### 步骤 4：验证配置正确性

检查 YAML 语法是否正确，确保所有路径修改到位。

## 预期结果

生成的 `docker-compose-dev.yaml` 将包含：
- 完整的三个服务配置（mysql、redis、app）
- 所有卷路径指向 `~/Docker/nao-todo-server/` 目录
- 所有配置文件挂载路径更新为开发环境路径
- 网络配置保持不变

同时在 `~/Docker/nao-todo-server/` 目录下生成：
- 完整的目录结构
- 从 `.example` 目录复制的配置文件（去掉 `.example` 后缀）
- `.env` 环境变量文件

## 注意事项

1. Windows 系统下，Docker Desktop 需要配置正确的文件共享权限
2. 用户需要确保 Docker Desktop 已配置允许访问 `~/Docker/` 目录
3. 配置文件中的占位符（如 `YOUR_REDIS_PASSWORD_HERE`）需要用户手动替换

## 风险评估

- **低风险**：仅修改路径配置，不影响服务逻辑
- **潜在问题**：Docker Desktop 文件共享权限问题，可能需要用户手动配置

## 输出文件

- `docker-compose-dev.yaml` - 开发环境 Docker Compose 配置文件
- `~/Docker/nao-todo-server/mysql/conf/my.cnf` - MySQL 配置文件
- `~/Docker/nao-todo-server/mysql/conf/client.cnf` - MySQL 客户端配置文件
- `~/Docker/nao-todo-server/redis/conf/redis.conf` - Redis 配置文件
- `~/Docker/nao-todo-server/.env` - 环境变量文件
