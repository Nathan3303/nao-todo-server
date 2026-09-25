# Nao Todo Server 部署指南

> 适用于 Ubuntu 22.04 LTS 服务器，2核2G配置

---

## 目录

- [1. 环境准备](#1-环境准备)
- [2. 安装 Docker 和 Docker Compose](#2-安装-docker-和-docker-compose)
- [3. 创建项目目录结构](#3-创建项目目录结构)
- [4. 克隆项目代码](#4-克隆项目代码)
- [5. 配置环境变量](#5-配置环境变量)
- [6. 创建数据库配置文件](#6-创建数据库配置文件)
- [7. 启动服务](#7-启动服务)
- [8. 验证服务](#8-验证服务)
- [9. 防火墙配置](#9-防火墙配置)
- [10. 常用管理命令](#10-常用管理命令)
- [11. 安全建议](#11-安全建议)
- [12. 故障排除](#12-故障排除)

---

## 1. 环境准备

### 1.1 服务器要求

| 项目 | 要求               |
| ---- | ------------------ |
| CPU  | 2核                |
| 内存 | 2GB                |
| 磁盘 | 至少 10GB 可用空间 |
| 系统 | Ubuntu 22.04 LTS   |
| 网络 | 开放 3302 端口     |

### 1.2 登录服务器

```bash
ssh root@your-server-ip
```

### 1.3 更新系统与安装基础依赖

```bash
# 更新系统软件包
apt update && apt upgrade -y

# 安装基础工具
apt install -y curl wget git vim unzip apt-transport-https ca-certificates software-properties-common
```

---

## 2. 安装 Docker 和 Docker Compose

```bash
# 添加 Docker GPG 密钥
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# 添加 Docker 仓库
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

# 更新并安装 Docker
apt update && apt install -y docker-ce docker-ce-cli containerd.io

# 安装 Docker Compose
apt install -y docker-compose-plugin

# 验证安装
docker --version
docker compose version
```

### 2.1 配置 Docker 权限（可选）

```bash
# 将当前用户加入 docker 组
usermod -aG docker $USER
newgrp docker
```

---

## 3. 创建项目目录结构

```bash
# 创建项目根目录
mkdir -p /opt/nao-todo-server
cd /opt/nao-todo-server

# 创建数据和配置目录
mkdir -p deploy/data/mysql
mkdir -p deploy/data/redis
mkdir -p deploy/conf/mysql
mkdir -p deploy/conf/redis
mkdir -p logs
mkdir -p uploads/avatars
```

---

## 4. 克隆项目代码

```bash
git clone https://github.com/your-username/nao-todo-server.git .
```

---

## 5. 配置环境变量

创建 `.env` 文件：

```bash
cat > .env <<EOF
# JWT 密钥（生产环境请使用至少32位的复杂密钥）
JWT_SECRET=your-strong-jwt-secret-here-32chars-min

# 时区
TZ=Asia/Shanghai

# MySQL 配置
MYSQL_HOST=mysql
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PASSWORD=your-mysql-password-123
MYSQL_DATABASE=naotodo

# Redis 配置
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password-456
REDIS_DB=0
EOF
```

> **重要**：请务必修改上述密码为安全的随机密码

---

## 6. 创建数据库配置文件

### 6.1 MySQL 配置

创建 `deploy/conf/mysql/my.cnf`：

```bash
cat > deploy/conf/mysql/my.cnf <<EOF
[mysqld]
character-set-server=utf8mb4
collation-server=utf8mb4_unicode_ci
init_connect='SET NAMES utf8mb4'
max_allowed_packet=64M
wait_timeout=86400
interactive_timeout=86400
bind-address=0.0.0.0

[mysql]
default-character-set=utf8mb4

[client]
default-character-set=utf8mb4
EOF
```

创建 `deploy/conf/mysql/client.cnf`：

```bash
cat > deploy/conf/mysql/client.cnf <<EOF
[client]
default-character-set=utf8mb4
EOF
```

### 6.2 Redis 配置

创建 `deploy/conf/redis/redis.conf`：

```bash
cat > deploy/conf/redis/redis.conf <<EOF
bind 0.0.0.0
port 6379
requirepass your-redis-password-456
maxmemory 512mb
maxmemory-policy allkeys-lru
appendonly yes
save 900 1
save 300 10
save 60 10000
EOF
```

> **注意**：Redis 密码需与 `.env` 中的 `REDIS_PASSWORD` 保持一致

---

## 7. 启动服务

```bash
# 构建并启动所有容器
docker compose up -d

# 查看启动日志
docker compose logs -f
```

---

## 8. 验证服务

### 8.1 查看容器状态

```bash
docker compose ps
```

### 8.2 测试 API

```bash
# 测试健康检查接口
curl http://localhost:3302/api/health
```

### 8.3 预期输出

```json
{
    "code": 200,
    "message": "success",
    "data": {
        "status": "healthy"
    }
}
```

---

## 9. 防火墙配置

```bash
# 查看当前状态
ufw status

# 允许 SSH 访问
ufw allow ssh

# 允许应用端口（3302）
ufw allow 3302/tcp

# 禁止外部访问数据库端口（重要）
ufw deny 3306/tcp
ufw deny 6379/tcp

# 启用防火墙
ufw enable

# 查看规则
ufw status verbose
```

---

## 10. 常用管理命令

```bash
# 启动服务
docker compose up -d

# 停止服务
docker compose down

# 查看日志
docker compose logs -f

# 查看容器状态
docker compose ps

# 重启单个服务
docker compose restart app

# 进入 MySQL 容器
docker compose exec mysql mysql -u root -p

# 进入 Redis 容器
docker compose exec redis redis-cli -a your-redis-password-456

# 查看 Docker 资源使用
docker stats

# 更新代码后重启
git pull && docker compose up -d --build
```

---

## 11. 安全建议

### 11.1 基础安全

1. **修改默认密码**：务必更新 `.env` 中的所有密码
2. **禁用 root 远程登录**：修改 `/etc/ssh/sshd_config`
3. **使用非 root 用户**：创建普通用户进行日常操作

### 11.2 Nginx 反向代理（推荐）

```bash
# 安装 Nginx
apt install -y nginx

# 创建配置文件
cat > /etc/nginx/sites-available/naotodo <<EOF
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:3302;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
    }
}
EOF

# 启用配置
ln -s /etc/nginx/sites-available/naotodo /etc/nginx/sites-enabled/

# 测试并重启
nginx -t
systemctl restart nginx
```

### 11.3 SSL 证书配置

```bash
# 安装 Certbot
apt install -y certbot python3-certbot-nginx

# 申请证书
certbot --nginx -d your-domain.com
```

---

## 12. 故障排除

### 12.1 Docker 启动失败

```bash
# 查看 Docker 状态
systemctl status docker

# 查看 Docker 日志
journalctl -u docker
```

### 12.2 MySQL 连接失败

```bash
# 检查 MySQL 容器是否正常运行
docker compose ps mysql

# 进入 MySQL 容器检查
docker compose exec mysql mysql -u root -p

# 检查数据库是否存在
SHOW DATABASES;
```

### 12.3 内存不足

```bash
# 查看内存使用
free -h

# 检查 Redis 内存配置
docker compose exec redis redis-cli -a your-redis-password-456 INFO memory
```

### 12.4 端口占用

```bash
# 检查端口占用情况
netstat -tlnp | grep 3302
netstat -tlnp | grep 3306
netstat -tlnp | grep 6379
```

---

## 附录：项目目录结构

```
/opt/nao-todo-server/
├── deploy/
│   ├── conf/
│   │   ├── mysql/
│   │   │   ├── my.cnf
│   │   │   └── client.cnf
│   │   └── redis/
│   │       └── redis.conf
│   └── data/
│       ├── mysql/          # MySQL 数据目录
│       └── redis/          # Redis 数据目录
├── logs/                   # 应用日志
├── uploads/
│   └── avatars/            # 用户头像上传目录
├── .env                    # 环境变量配置
├── docker-compose.yml      # Docker Compose 配置
├── Dockerfile              # 应用 Docker 镜像构建文件
└── ...                     # 项目代码
```

---

## 附录：配置说明

### 环境变量说明

| 变量             | 说明             | 默认值        |
| ---------------- | ---------------- | ------------- |
| `JWT_SECRET`     | JWT 签名密钥     | 必须设置      |
| `MYSQL_PASSWORD` | MySQL 密码       | 必须设置      |
| `MYSQL_DATABASE` | 数据库名称       | naotodo       |
| `REDIS_PASSWORD` | Redis 密码       | 必须设置      |
| `REDIS_DB`       | Redis 数据库编号 | 0             |
| `TZ`             | 时区             | Asia/Shanghai |

### 端口说明

| 端口 | 服务     | 是否对外开放     |
| ---- | -------- | ---------------- |
| 3302 | 应用服务 | 是               |
| 3306 | MySQL    | 否（防火墙禁止） |
| 6379 | Redis    | 否（防火墙禁止） |

---

**部署完成！** 服务已在 `http://your-server-ip:3302` 运行。
