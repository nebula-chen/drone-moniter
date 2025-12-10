drone-stats-service

简介
---
`drone-stats-service` 负责无人机飞行数据的统计、查询与导出，依赖 InfluxDB （时序数据）和 MySQL（元数据/导出记录）。

主要功能
- 时序统计（时间序列查询、聚合）
- 导出飞行记录（同步/异步）
- 近期轨迹查询、分时统计与负载统计

重要文件
- 配置：`etc/dronestats.yaml`
- 服务入口：`dronestats.go`
- Docker：`Dockerfile`、`docker-compose.yml`

快速开始（依赖服务）
- 需要先启动 `influxDB`（仓库内的 `influxDB/`），以及 MySQL（若 docker-compose 中包含）。

快速开始（推荐）
1. 使用仓库根目录的一键脚本：
	```bash
	./deploy.sh
	```
	脚本会启动 `influxDB`、`drone-api`、`drone-stats-service`。

2. 或单独启动：
```bash
cd influxDB
docker-compose up -d --build

cd ../drone-stats-service
docker-compose up --build -d
```

配置与端口
- 服务架构：`drone-stats-service/drone_stats.api`   // 包含结构体定义、接口定义及路由配置等
- 配置文件：`drone-stats-service/etc/dronestats.yaml`（调整 InfluxDB/DB 连接和端口）
- 默认对外端口（仓库约定）：8088（nginx 示例通过该端口做反向代理到 `/record`）

日志与数据目录
- 日志：`drone-stats-service/log/`
- MySQL 持久化数据示例：`drone-stats-service/mysql_data/`
- InfluxDB 数据位于 `influxDB/influxdb2/`（仓库中的示例数据目录）

提示
- 启动顺序：先启动 InfluxDB -> 再启动 `drone-stats-service`（或使用 `deploy.sh` 自动序列化启动）。
- 导出任务可能会写入数据库或本地文件，查看 `internal/export` 目录了解任务实现。

