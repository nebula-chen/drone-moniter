package svc

import (
	"autonomous-vehicle/internal/config"
	"autonomous-vehicle/internal/dao"
	"autonomous-vehicle/internal/websocket"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"

	"crypto/sha1"
	"encoding/hex"
	"math/rand"
	"sort"
	"strconv"
)

// ServiceContext is the context for the service
type ServiceContext struct {
	Config   config.Config
	WSHub    *websocket.Hub
	Dao      *dao.InfluxDao
	MySQLDao *dao.MySQLDao

	OnlineDrones sync.Map // key: uasID, value: time.Time

	Token      string     // 新增：缓存token
	TokenMutex sync.Mutex // 新增：token并发保护
}

// NewServiceContext creates a new service context
func NewServiceContext(c config.Config) *ServiceContext {
	hub := websocket.NewHub()
	go hub.Run()
	URL := "http://" + c.InfluxDBConfig.Host + ":" + c.InfluxDBConfig.Port
	options := influxdb2.DefaultOptions().
		SetBatchSize(c.InfluxDBConfig.BatchSize).               // 批量大小
		SetFlushInterval(c.InfluxDBConfig.FlushInterval * 1000) // 毫秒
		// SetPrecision(time.Second)
	client := influxdb2.NewClientWithOptions(URL, c.InfluxDBConfig.Token, options)

	_, err := client.Ping(context.Background())
	if err != nil {
		panic("InfluxDB connect error: " + err.Error())
	}
	ctx := &ServiceContext{
		Config: c,
		WSHub:  hub,
		Dao:    dao.NewInfluxDao(client, c.InfluxDBConfig.Org, c.InfluxDBConfig.Bucket),
	}

	// 初始化 MySQL（如果配置了 DataSource）
	if c.MySQL.DataSource != "" {
		mdao, err := dao.NewMySQLDao(c.MySQL.DataSource)
		if err != nil {
			panic("MySQL connect error: " + err.Error())
		}
		ctx.MySQLDao = mdao
	}

	// 启动定时清理协程
	go func() {
		for {
			now := time.Now()
			ctx.OnlineDrones.Range(func(key, value interface{}) bool {
				lastTime, ok := value.(time.Time)
				if ok && now.Sub(lastTime) > time.Minute {
					ctx.OnlineDrones.Delete(key)
				}
				return true
			})
			time.Sleep(10 * time.Second)
		}
	}()

	return ctx
}

// 获取token
func (ctx *ServiceContext) GetToken() (string, error) {
	ctx.TokenMutex.Lock()
	defer ctx.TokenMutex.Unlock()
	if ctx.Token != "" {
		return ctx.Token, nil
	}
	// 获取token
	url := fmt.Sprintf("%s?grant_type=client_credentials&client_id=%s&client_secret=%s",
		ctx.Config.TokenURL, ctx.Config.ClientID, ctx.Config.ClientSecret)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	ctx.Token = result.AccessToken
	return ctx.Token, nil
}

// 生成两位随机数
func (ctx *ServiceContext) GenNonce() string {
	return strconv.Itoa(rand.Intn(90) + 10) // 10~99
}

// 生成签名
func (ctx *ServiceContext) GenSignature(timestamp, nonce string) string {
	appSecret := ctx.Config.ClientSecret
	params := []string{appSecret, timestamp, nonce}
	sort.Strings(params)
	content := params[0] + params[1] + params[2]
	h := sha1.New()
	h.Write([]byte(content))
	return hex.EncodeToString(h.Sum(nil))
}

// 生成签名相关参数（时间戳、nonce、signature、access_token）
func (ctx *ServiceContext) GenSignParams() (timestamp, nonce, signature, accessToken string, err error) {
	timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	nonce = ctx.GenNonce()
	signature = ctx.GenSignature(timestamp, nonce)
	accessToken, err = ctx.GetToken()
	return
}
