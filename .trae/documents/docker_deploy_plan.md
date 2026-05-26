# Docker Compose 一键部署方案

## 项目分析

### 当前架构
- **后端**: Golang (Go 1.20+)
- **数据库**: MySQL 8.0+（外部容器）
- **缓存**: Redis 7.0+
- **端口**: 3302

### 配置现状
当前配置文件 `conf/config.yaml` 使用本地连接：
- MySQL: `localhost:3306`
- Redis: `localhost:6379`

## 部署方案

### 1. 创建文件清单

| 文件 | 说明 | 作用 |
|------|------|------|
| `Dockerfile` | 后端服务镜像构建文件 | 构建 Go 应用镜像 |
| `docker-compose.yml` | 容器编排配置 | 启动后端和 Redis |
| `.env` | 环境变量配置 | 存储敏感配置 |

### 2. MySQL 外部容器部署

#### 步骤 1: 创建 MySQL 数据目录
```bash
mkdir -p /home/nathan/Docker/mysql3/data
mkdir -p /home/nathan/Docker/mysql3/conf.d
mkdir -p /home/nathan/Docker/mysql3/backup
```

#### 步骤 2: 创建 .env 文件
```bash
echo "MYSQL_ROOT_PASSWORD=lianGjh3303.." > /home/nathan/Docker/mysql3/.env
echo "TZ=Asia/Shanghai" >> /home/nathan/Docker/mysql3/.env
```

#### 步骤 3: 创建 client.cnf（可选）
```bash
cat > /home/nathan/Docker/mysql3/client.cnf <<EOF
[client]
default-character-set=utf8mb4
EOF
```

#### 步骤 4: 启动 MySQL 容器
```bash
sudo docker run -d --name mysql3 \
  -p 3306:3306 \
  -v /home/nathan/Docker/mysql3/data:/var/lib/mysql \
  -v /home/nathan/Docker/mysql3/conf.d:/etc/mysql/conf.d \
  -v /home/nathan/Docker/mysql3/backup:/backup \
  -v /home/nathan/Docker/mysql3/client.cnf:/etc/mysql/client.cnf:ro \
  --env-file /home/nathan/Docker/mysql3/.env \
  mysql:latest
```

#### 步骤 5: 创建数据库
```bash
docker exec -it mysql3 mysql -u root -p
# 输入密码后执行
CREATE DATABASE IF NOT EXISTS naotodo_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. Docker Compose 配置

#### 网络配置
- Redis 服务名: `redis`
- 后端服务通过主机网络访问 MySQL

#### 数据持久化
- Redis 数据挂载到 `./data/redis`

### 4. 服务依赖顺序
1. MySQL 外部容器已启动
2. Redis 启动
3. 后端服务启动（依赖 MySQL 和 Redis）

## 实施步骤

### 步骤 1: 创建 Dockerfile
- 使用多阶段构建（build + runtime）
- 构建阶段: golang:1.22-alpine
- 运行阶段: alpine:latest

### 步骤 2: 创建 docker-compose.yml
- 定义 redis 服务（端口 6379）
- 定义 app 服务（端口 3302）
- 使用 host.docker.internal 访问主机 MySQL

### 步骤 3: 创建 .env 文件
- 设置数据库密码
- 设置 JWT 密钥
- 设置 Redis 密码

### 步骤 4: 更新配置支持环境变量
- 修改 `conf/config.go` 支持环境变量覆盖
- MySQL 连接地址使用 `host.docker.internal`（Docker Desktop）或主机 IP

## 部署命令

```bash
# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据
docker-compose down -v
```

## 注意事项

1. **端口冲突**: 确保主机 3302、6379 端口未被占用
2. **MySQL 网络**: 
   - Docker Desktop: 使用 `host.docker.internal` 访问主机 MySQL
   - Linux: 需要配置容器网络或使用 `--network=host`
3. **数据备份**: 定期备份 `/home/nathan/Docker/mysql3/data` 和 `./data` 目录
4. **安全配置**: 生产环境需修改默认密码
5. **防火墙**: 云服务器需开放 3302 端口