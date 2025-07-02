package main

import (
	"context"
	"database/sql"
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

	ctx := svc.NewServiceContext(c)
	// 自动建表
	if err := autoMigrate(ctx.MySQLDao.DB); err != nil {
		panic(err)
	}

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)

	// 启动定时任务
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // 每15秒拉取一次
		defer ticker.Stop()
		fmt.Println("开始拉取数据...")
		for {
			processAllUasData(ctx)
			<-ticker.C
		}
	}()

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	processAllUasData(ctx)
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
		if err := ctx.MySQLDao.RegisterSortiesIfNotExist(id, regTime); err != nil {
			fmt.Println("注册无人机失败:", id, err)
		}

		// 拉取该无人机近一段时间的飞行数据并处理
		end := time.Now().UTC()
		start := time.Date(2025, 6, 19, 11, 0, 0, 0, time.UTC) // 从25年6月19号19点整开始拉取数据（丰翼数据上报接口当天19:20发布生产环境）
		// start := end.Add(-1 * time.Hour) // 只拉取最近1小时

		// 复用已有逻辑
		req := &types.FlightRecordReq{
			OrderID:   id,
			StartTime: start.Format(time.RFC3339),
			EndTime:   end.Format(time.RFC3339),
		}
		logic := logic.NewGetFlightRecordsLogic(context.Background(), ctx)
		_, err := logic.GetFlightRecords(req)
		if err != nil {
			fmt.Println("处理无人机飞行数据失败:", id, err)
		}
	}
}

func autoMigrate(db *sql.DB) error {
	// flight_sorties 表
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS flight_sorties (
        id INT AUTO_INCREMENT PRIMARY KEY,
        OrderID VARCHAR(128) NOT NULL UNIQUE,
        register_time DATETIME,
        model VARCHAR(64)
    );`)
	if err != nil {
		return err
	}

	// flight_records 表
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS flight_records (
        id INT AUTO_INCREMENT PRIMARY KEY,
        OrderID VARCHAR(128) NOT NULL,
        start_time DATETIME NOT NULL,
        end_time DATETIME,
        start_lat BIGINT,
        start_lng BIGINT,
        end_lat BIGINT,
        end_lng BIGINT,
        distance DOUBLE(10,2),
        battery_used INT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		payload INT NOT NULL DEFAULT 0
    );`)
	if err != nil {
		return err
	}

	// flight_track_points 表
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS flight_track_points (
        id INT AUTO_INCREMENT PRIMARY KEY,
        OrderID VARCHAR(128) NOT NULL,
        flightStatus VARCHAR(16),
        timeStamp DATETIME,
        longitude BIGINT,
        latitude BIGINT,
        heightType INT,
        height INT,
        altitude INT,
        VS INT,
        GS INT,
        course INT,
        SOC INT,
        RM INT,
        windSpeed INT,
        windDirect INT,
        temperture INT,
        humidity INT
    );`)
	return err
}
