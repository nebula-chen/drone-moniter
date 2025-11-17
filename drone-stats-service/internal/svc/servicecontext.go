package svc

import (
	"drone-stats-service/internal/config"
	"drone-stats-service/internal/dao"
	"drone-stats-service/internal/export"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type ServiceContext struct {
	Config      config.Config
	InfluxDao   *dao.InfluxDao
	MySQLDao    *dao.MySQLDao
	TaskManager *export.TaskManager
}

func NewServiceContext(c config.Config) *ServiceContext {
	influxClient := influxdb2.NewClient("http://"+c.InfluxDBConfig.Host+":"+c.InfluxDBConfig.Port, c.InfluxDBConfig.Token)
	influxDao := dao.NewInfluxDao(influxClient, c.InfluxDBConfig.Org)
	mysqlDao, err := dao.NewMySQLDao(c.MySQL.DataSource)
	if err != nil {
		panic(err)
	}
	// 初始化 TaskManager，使用系统临时目录存放任务及输出
	baseURL := fmt.Sprintf("http://%s:%d", c.Host, c.Port)
	taskMgr, _ := export.NewTaskManager(mysqlDao, influxDao, "", baseURL)
	return &ServiceContext{
		Config:      c,
		InfluxDao:   influxDao,
		MySQLDao:    mysqlDao,
		TaskManager: taskMgr,
	}
}
