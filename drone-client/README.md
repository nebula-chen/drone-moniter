# drone-client

简介
---
`drone-client` 是一个轻量的无人机模拟上报程序，用于向后端（如 `drone-api`）周期性发送无人机状态数据，方便本地联调、功能验证与演示。

主要功能
- 模拟无人机的起飞、爬升、巡航、下降与降落流程
- 根据速度与航向计算轨迹点（见 `flight/GetNewLatLon`）
- 以 JSON 格式通过 HTTP POST 向后端上报位置信息和状态

源码要点
- 程序入口：`client.go`
- 轨迹计算：`flight/flight.go`（球面几何计算，返回新的经/纬度）
- 上报接口默认地址在 `client.go` 中的 `serverURL` 变量（请根据实际部署修改）

快速开始（开发环境）
1. 安装 Go（建议 >= 1.18），并启用模块支持：
   ```bash
   cd drone-client
   go run client.go
   ```
2. 运行时可以传入命令行参数来覆盖默认值（在 `client.go` 中使用 `flag` 定义）：
   - `-lat` 初始纬度（默认示例：22.8007210）
   - `-lon` 初始经度（默认示例：113.9530990）
   - `-bearing` 初始航向角（度）
   - `-id` 无人机 ID

示例：
```bash
cd drone-client
# 使用自定义初始位置与 ID
go run client.go -lat=22.8010 -lon=113.9540 -bearing=90 -id=uas-demo-1
```

说明：默认程序把经纬度乘以 1e7 转为整数后上报（见 `client.go` 中的 `Longitude`/`Latitude` 字段处理），高度/速度也做了相应缩放，请与后端约定的单位保持一致。

修改上报地址
---
当前示例代码中上报的目标 URL 写在 `client.go` 的 `serverURL` 变量中：
```go
serverURL := "http://localhost:19999/api/drone/status"
```
请修改为你的 `drone-api` 地址，或在启动前用文本编辑器替换。

构建可执行文件
```bash
go build -o drone-client ./client.go
./drone-client -id=uas1
```

日志与输出
- 程序会在控制台打印当前经纬度、高度、电量和阶段信息，便于观察模拟过程。

注意事项与调试提示
- 确保后端接口可达且能够接受 JSON POST（Content-Type: application/json）。
- 若使用反向代理（如 nginx），确保路径与代理规则匹配（仓库根目录 `nginx_drone.conf` 提供示例）。
- 若你希望批量模拟多台无人机，可复制并启动多个进程，或修改代码以并行生成多条上报流。

扩展建议
- 支持从配置文件加载目标地址与多个无人机的初始参数
- 增加 WebSocket 发送或基于 gRPC 的上报模式以适配不同后端
