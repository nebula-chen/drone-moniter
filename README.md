# drone-moniter
无人机监控系统

本仓库包含若干后端服务，用于无人机状态上报、统计与展示，包括：
- `drone-api`：实时无人机状态与 WebSocket 接入（API 网关）
- `drone-stats-service`：统计、查询、导出（依赖 InfluxDB / MySQL）
- `autonomous-vehicle`：车辆/设备信息服务与 WebSocket
- `influxDB`：InfluxDB 数据存储（用于统计服务）
- `ui`：前端静态资源（可部署到 nginx）

* 安装 nginx
```Shell
# 包管理器安装
sudo apt-get install nginx

配置文件目录：/etc/nginx/
默认网站根目录：/var/www/html/drone-moniter/
日志文件：/var/log/nginx/

```

* 安装 docker-compose
```Shell
# 下载最新稳定版 Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.36.1/docker-compose-$(uname -s)-$(uname -m)" -o docker-compose
sudo mv docker-compose /usr/local/bin/
# 赋予可执行权限
sudo chmod +x /usr/local/bin/docker-compose
# 验证安装
docker-compose --version
```

* 部署
```Shell
# 如果前端 html 页面或后端服务接口有修改则需要同步更新 nginx.conf 文件的配置
# 把 index.html 移动到 nginx 默认网站根目录
# 将无人机后端的nignx配置复制到 nginx 配置目录下
sudo cp -v ./nginx_drone.conf /etc/nginx/sites-enabled/
sudo nginx -t
sudo nginx -s reload
sudo systemctl start nginx
sudo mkdir /var/www/html/drone-moniter
sudo cp ./ui/* /var/www/html/drone-moniter


服务配置 & 端口（仓库约定／示例）
- `drone-api`：HTTP + WS，代理端口示例 19999（详见 `drone-api/etc/drone-api.yaml`）
- `drone-stats-service`：统计服务，代理端口示例 8088（详见 `drone-stats-service/etc/dronestats.yaml`），依赖 InfluxDB
- `autonomous-vehicle`：车辆信息服务，代理端口示例 8060（详见 `autonomous-vehicle/etc/autonomousvehicle.yaml`）

常用操作（在仓库根目录运行）
```bash
# 启动（按 deploy.sh 顺序自动启动依赖）
./deploy.sh

# 停止
./stop.sh

# 单个服务（示例）
cd drone-api && docker-compose up --build -d
cd ../drone-stats-service && docker-compose up --build -d
cd ../autonomous-vehicle && docker-compose up --build -d
```

更多说明
- 各服务下面都包含各自的 `README.md`（见子目录），包含更细的启动、配置与端点说明。
- 配置文件位于各服务的 `etc/` 目录（yaml 格式），请根据环境调整。

