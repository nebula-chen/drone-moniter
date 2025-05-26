package svc

import (
	"drone-api/internal/config"
	"drone-api/internal/dao"
	"drone-api/internal/websocket"
)

type ServiceContext struct {
	Config config.Config
	WSHub  *websocket.Hub
	Dao    *dao.InfluxDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	hub := websocket.NewHub()
	go hub.Run()
	// URL := "http://" + c.InfluxDBConfig.Host + ":" + c.InfluxDBConfig.Port
	// options := influxdb2.DefaultOptions().
	// 	SetBatchSize(c.InfluxDBConfig.BatchSize).               // 批量大小
	// 	SetFlushInterval(c.InfluxDBConfig.FlushInterval * 1000) // 毫秒
	// 	// SetPrecision(time.Second)
	// client := influxdb2.NewClientWithOptions(URL, c.InfluxDBConfig.Token, options)

	// _, err := client.Ping(context.Background())
	// if err != nil {
	// 	panic("InfluxDB connect error: " + err.Error())
	// }
	return &ServiceContext{
		Config: c,
		WSHub:  hub,
		// Dao:    dao.NewInfluxDao(nil, c.InfluxDBConfig.Org, c.InfluxDBConfig.Bucket),
	}
}
