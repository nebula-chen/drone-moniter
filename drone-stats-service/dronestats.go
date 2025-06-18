package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"drone-stats-service/internal/config"
	"drone-stats-service/internal/handler"
	"drone-stats-service/internal/logic"
	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/dronestats.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	// 启动定时任务
	go func() {
		ticker := time.NewTicker(15 * time.Second) // 每15秒拉取一次
		defer ticker.Stop()
		for {
			processAllUasData(ctx)
			<-ticker.C
		}
	}()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

func processAllUasData(ctx *svc.ServiceContext) {
	ids, err := ctx.InfluxDao.GetAllUasIDsAndFirstSeen()
	if err != nil {
		fmt.Println("拉取无人机ID失败:", err)
		return
	}
	for id, regTime := range ids {
		// 自动注册
		if err := ctx.MySQLDao.RegisterUasIfNotExist(id, regTime); err != nil {
			fmt.Println("注册无人机失败:", id, err)
		}
		// 拉取该无人机近一段时间的飞行数据并处理
		end := time.Now().UTC()
		start := end.Add(-24 * time.Hour) // 例如只处理最近24小时
		// 复用已有逻辑
		req := &types.FlightRecordReq{
			FlightCode: id,
			StartTime:  start.Format(time.RFC3339),
			EndTime:    end.Format(time.RFC3339),
		}
		logic := logic.NewGetFlightRecordsLogic(context.Background(), ctx)
		_, err := logic.GetFlightRecords(req)
		if err != nil {
			fmt.Println("处理无人机飞行数据失败:", id, err)
		}
	}
}
