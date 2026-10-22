# 服务端 CORS 放行「预览域 / 局域网」— 部署与验证说明（#52 / DEF-80）

> 代码：`interfaces/routers/cors.go`（判定唯一来源）+ `routers.go`（挂载）。单测：`cors_test.go`。
> PRD：`nao-todo/docs/prds/2026-10-22-server-cors-origins.md`。

## 一、放行规则（三组，可枚举）

| 组 | 规则 | 生效条件 |
| --- | --- | --- |
| 精确白名单 | `http://localhost:5173` · `http://localhost:5174` · `https://todo.nathanao.space` · `https://todobe.nathanao.space` | 始终 |
| 本项目预览域 | `^https://naotodo-[a-z0-9-]+-nathan3303s-projects\.vercel\.app$` | 始终（与 debug 无关） |
| 局域网 | `^http://<私网 IPv4>:(5173\|5174)$`，私网 = `192.168/16`、`10/8`、`172.16–31/12` | 仅 `APP_DEBUG=true` |

要点：

- `AllowCredentials: true` ⇒ 中间件**回显请求的精确 origin**，任何情况下不返回 `*`。
- 前两组为**常量比较 / 锚定整串正则**，`…vercel.app.evil.com`、`evil-vercel.app`、越界端口一律不命中。
- 不改 `AllowMethods` / `AllowHeaders` / `ExposeHeaders`；现有正式域放行能力只增不减。

## 二、本地重建与三组预检

```bash
cd /home/nathan/Projects/nao-todo-server
docker compose up -d --build          # 重建并启动（注意：不要提交 docker-compose.yml 的本地改动）
docker compose logs -f server | head  # 确认无启动错误
```

> `APP_DEBUG=true` 时监听 `3302`（见 `.env`）；生产镜像不设该变量，局域网规则不生效。

三组预检（被放行来源须逐字回显 `access-control-allow-origin`）：

```bash
API=http://localhost:3302/api/ping

# 1. 既有正式域（AC1，本地亦应放行）
curl -si -X OPTIONS "$API" \
  -H 'Origin: https://todobe.nathanao.space' \
  -H 'Access-Control-Request-Method: POST' | grep -i 'access-control'

# 2. 本项目预览域（AC3，debug 无关）
curl -si -X OPTIONS "$API" \
  -H 'Origin: https://naotodo-git-feat-52-cors-nathan3303s-projects.vercel.app' \
  -H 'Access-Control-Request-Method: POST' | grep -i 'access-control'

# 3. 局域网（AC3，仅 APP_DEBUG=true 放行；把 IP 换成你的局域网地址）
curl -si -X OPTIONS "$API" \
  -H 'Origin: http://192.168.2.211:5173' \
  -H 'Access-Control-Request-Method: POST' | grep -i 'access-control'
```

负例（应**完全没有** `access-control-allow-origin`）：

```bash
for o in 'https://naotodo-x-nathan3303s-projects.vercel.app.evil.com' \
         'https://evil-vercel.app' \
         'http://192.168.2.211:9999'; do
  echo "== $o"
  curl -si -X OPTIONS "$API" -H "Origin: $o" -H 'Access-Control-Request-Method: POST' \
    | grep -i 'access-control-allow-origin' || echo '无 ACAO（正确）'
done
```

生产环境（`APP_DEBUG` 未设）额外复核：局域网来源必须**无** ACAO，正式域与预览域正常回显。

## 三、回滚

`git revert <squash commit>` + 重新部署。CORS 纯配置，无数据迁移。
