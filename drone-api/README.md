drone-api

简介
---
`drone-api` 提供无人机实时状态接入与 WebSocket 转发，适合作为前端地图/控制台的实时数据源。

主要功能
- 接收无人机状态上报（HTTP/UDP/自定义）
- 状态数据写入到influxDB
- 提供实时 WebSocket 推送
- 在线数量统计等轻量接口

重要文件
- 配置：`etc/drone-api.yaml`
- 服务入口：`drone.go`
- Docker：`Dockerfile`、`docker-compose.yml`
- 静态资源：`resources/`（可部署到 nginx）

快速开始（开发）
```bash
cd drone-api
go run drone.go
```

快速开始（Docker Compose）
```bash
cd drone-api
docker-compose up --build -d
```

配置与端口
- 服务架构：`drone-api/drone.api`   // 包含结构体定义、接口定义及路由配置等
- 配置文件：`drone-api/etc/drone-api.yaml`（调整绑定端口、数据库/缓存等）
- 默认对外端口（仓库约定）：19999（nginx 示例通过该端口做反向代理）

常见端点（示例）
- HTTP 接口：`/api/` 下的各类 REST 接口（详见 `internal/handler/routes.go`）
- WebSocket：`/api/ws`（用于前端建立实时连接）

日志与数据
- 日志目录：`drone-api/log/`
- 持久化或缓存（如需要）在配置文件中指定

提示
- 若使用 nginx 代理，请参考仓库根目录的 `nginx_drone.conf`。
- 前端静态文件位于 `drone-api/resources/`，可以直接拷贝到 nginx 的站点目录。

