drone-stats-service

简介
---
`drone-stats-service` 负责无人机飞行数据的采集统计、查询与导出，依赖 InfluxDB（时序数据）和 MySQL（元数据/导出记录）。

主要功能
- **时序统计**: 多维度时间序列查询与聚合（支持分时/聚合窗口）。
- **导出飞行记录**: 支持同步与异步导出，导出任务可写入数据库或文件。
- **近期轨迹与分段统计**: 查询近段轨迹、分时统计与负载/使用率分析。
- **API 与导出管理**: 提供 REST 接口和任务队列，用于调度导出与备份。

重要文件
- **配置**: `etc/dronestats.yaml`（服务配置、数据库/Influx 连接等）
- **服务入口**: `dronestats.go`
- **接口与路由**: `drone_stats.api`（结构体定义、接口描述与路由）
- **Docker**: `Dockerfile`、`docker-compose.yml`
- **导出实现**: `internal/export/exporter.go`、`internal/export/taskmanager.go`
- **数据库访问**: `internal/dao/influx.go`、`internal/dao/mysql.go`
- **主要逻辑与处理器**: `internal/handler/`、`internal/logic/`

快速开始（依赖服务）
- 先启动依赖的时序/元数据服务：
	- `influxDB`（仓库根目录下的 `influxDB/`）
	- MySQL（如果使用仓库提供的 `docker-compose`，会在 `drone-stats-service` 服务定义中包含）

快速开始（推荐）
1. 使用仓库根目录的一键脚本（推荐按序启动所有服务）：
```bash
./deploy.sh
```
脚本会依次启动 `influxDB`、`drone-api`、`drone-stats-service` 等。

2. 单独启动（仅本服务）
```bash
cd influxDB
docker-compose up -d --build

cd ../drone-stats-service
docker-compose up --build -d
```

配置与端口
- **配置文件**: `drone-stats-service/etc/dronestats.yaml`（调整 InfluxDB、MySQL、端口和导出设置）
- **默认对外端口**: `8088`（仓库约定；示例 nginx 在 `nginx_drone.conf` 中把路径代理到 `/record`）

日志与数据目录
- **日志**: `drone-stats-service/log/`
- **MySQL 数据**: `drone-stats-service/mysql_data/`
- **InfluxDB 数据**: `influxDB/influxdb2/`

导出任务与实现细节
- 导出功能的主要实现位于 `internal/export/`，包含导出器和任务管理器：
	- `exporter.go`: 导出逻辑（支持多种导出目标）
	- `taskmanager.go`: 异步任务队列与调度
- 导出记录与任务状态通常会写入 MySQL（查看 `internal/dao/mysql.go` 与 `internal/export`）

开发与调试
- 本地运行（直接使用 Go）:
```bash
cd drone-stats-service
go run dronestats.go
```
- 使用 Docker 进行容器化运行:
```bash
cd drone-stats-service
docker-compose up --build -d
```

附加说明
- 启动顺序: 先启动 InfluxDB，再启动 `drone-stats-service`（或使用 `deploy.sh` 自动化）。
- 日志与导出异常: 查看 `log/` 目录下的最新日志以排查问题。
- 如果要查看 API 文档或前端演示，请参考 `resources/` 下的静态文件。

