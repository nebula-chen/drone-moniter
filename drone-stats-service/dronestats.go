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
		ticker := time.NewTicker(15 * time.Second) // 每15秒拉取一次
		defer ticker.Stop()
		fmt.Println("开始拉取数据...")
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

func autoMigrate(db *sql.DB) error {
	// uas_devices 表
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS uas_devices (
        id INT AUTO_INCREMENT PRIMARY KEY,
        uas_id VARCHAR(64) NOT NULL UNIQUE,
        register_time DATETIME,
        last_online_time DATETIME,
        status TINYINT,
        model VARCHAR(64)
    );`)
	if err != nil {
		return err
	}

	// flight_records 表
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS flight_records (
        id INT AUTO_INCREMENT PRIMARY KEY,
        uav_id VARCHAR(64) NOT NULL,
        start_time DATETIME NOT NULL,
        end_time DATETIME,
        start_lat BIGINT,
        start_lng BIGINT,
        end_lat BIGINT,
        end_lng BIGINT,
        distance DOUBLE(10,2),
        battery_used INT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );`)
	if err != nil {
		return err
	}

	// flight_track_points 表
	_, err = db.Exec(`
    CREATE TABLE IF NOT EXISTS flight_track_points (
        id INT AUTO_INCREMENT PRIMARY KEY,
        flight_record_id INT NOT NULL,
        flight_status VARCHAR(16),
        time_stamp DATETIME,
        longitude BIGINT,
        latitude BIGINT,
        altitude DOUBLE(6,1),
        soc INT
    );`)
	return err
}
