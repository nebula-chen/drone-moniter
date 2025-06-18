package svc

import (
	"drone-stats-service/internal/config"
	"drone-stats-service/internal/dao"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type ServiceContext struct {
	Config    config.Config
	InfluxDao *dao.InfluxDao
	MySQLDao  *dao.MySQLDao
}

func NewServiceContext(c config.Config) *ServiceContext {
	influxClient := influxdb2.NewClient("http://"+c.InfluxDBConfig.Host+":"+c.InfluxDBConfig.Port, c.InfluxDBConfig.Token)
	influxDao := dao.NewInfluxDao(influxClient, c.InfluxDBConfig.Org)
	mysqlDao, err := dao.NewMySQLDao(c.MySQL.DataSource)
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:    c,
		InfluxDao: influxDao,
		MySQLDao:  mysqlDao,
	}
}
