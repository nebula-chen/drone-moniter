autonomous-vehicle

简介
---
`autonomous-vehicle` 服务负责车辆/设备信息管理与 WebSocket 推送，适用于需要查询车辆列表、车辆详情、在线统计和规划路线等功能的场景。

主要功能
- 提供设备/车辆信息查询接口
- 提供在线设备统计接口
- 提供 WebSocket 实时推送（用于前端地图或控制台）
- 支持导出车辆记录

重要文件
- 配置：`etc/autonomousvehicle.yaml`
- 服务入口：`autonomousvehicle.go`
- Docker：`Dockerfile`、`docker-compose.yml`
- 日志目录示例：`log/`

快速开始（开发）
1. 先确保已安装 Go（>=1.18）并启用 Go Modules：
	```bash
	cd autonomous-vehicle
	go run autonomousvehicle.go
	```

快速开始（使用 Docker Compose）
```bash
cd autonomous-vehicle
docker-compose up --build -d
```

配置与端口
- 服务架构：`autonomous-vehicle/autonomousVehicle.api`   // 包含结构体定义、接口定义及路由配置等
- 配置文件：`autonomous-vehicle/etc/autonomousvehicle.yaml`（请根据环境修改数据库/端口等设置）
- 默认对外端口（仓库约定）：8060（nginx 中 proxy 示例使用该端口）

常见端点（示例）
- HTTP 接口：`/vehicle/`（列表、详情等，详见代码中的 `handler/routes.go`）
- WebSocket：`/vehicle/ws`（用于实时推送）

日志与数据
- 日志目录：`autonomous-vehicle/log/`
- 数据库连接配置在 `etc/autonomousvehicle.yaml` 中，常见为 MySQL

提示
- 本服务可以单独运行，也可配合 nginx 进行反向代理与统一入口管理。
- 如需在本地调试，请在 `etc/` 中修改为本地数据库连接并重启服务。
