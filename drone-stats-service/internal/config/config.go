package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	InfluxDBConfig InfluxDB
	MySQL          MySQLConf
	BackupConf     BackupConf
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

type BackupConf struct {
	BackupDir     string `json:"backupDir"`     // 备份存放目录（容器内路径，建议映射到宿主）
	IntervalDays  int    `json:"intervalDays"`  // 定期备份间隔（天）
	RetentionDays int    `json:"retentionDays"` // 备份保留天数
	InfluxBucket  string `json:"influxBucket"`  // 要导出的 InfluxDB bucket 名称
}
