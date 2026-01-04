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
	// 以下为可配置的重试与队列参数
	RetryMaxAttempts    int    `json:"retryMaxAttempts"`    // 最大重试次数
	RetryBaseDelayMs    int    `json:"retryBaseDelayMs"`    // 指数退避基准延迟（毫秒）
	ReplayerIntervalSec int    `json:"replayerIntervalSec"` // 后台重放间隔（秒）
	QueuePath           string `json:"queuePath"`           // 本地队列文件路径
}
