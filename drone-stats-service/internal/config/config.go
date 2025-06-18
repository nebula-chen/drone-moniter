package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	InfluxDBConfig InfluxDB
	MySQL          MySQLConf
}

type InfluxDB struct {
	Host            string
	Port            string
	User            string
	Password        string
	Token           string
	Bucket          string
	Org             string
	RetentionPolicy string
	Precision       string
	Timeout         string
	BatchSize       uint
	FlushInterval   uint
}

type MySQLConf struct {
	DataSource string
}
