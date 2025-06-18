package main

import (
	"bytes"
	"client/flight"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"time"
)

type DroneStatusReq struct {
	OrderID        string `json:"orderID"`
	FlightCode     string `json:"flightCode"`
	Sn             string `json:"sn"`
	FlightStatus   string `json:"flightStatus"`
	ManufacturerID string `json:"manufacturerID"`
	UasID          string `json:"uasID"`
	TimeStamp      string `json:"timeStamp"`
	UasModel       string `json:"uasModel"`
	Coordinate     int    `json:"coordinate"`
	Longitude      int64  `json:"longitude"`
	Latitude       int64  `json:"latitude"`
	HeightType     int    `json:"heightType"`
	Height         int    `json:"height"`
	Altitude       int    `json:"altitude"`
	VS             int    `json:"VS"`
	GS             int    `json:"GS"`
	Course         int    `json:"course"`
	SOC            int    `json:"SOC"`
	RM             int    `json:"RM"`
	WindSpeed      int    `json:"windSpeed"`
	WindDirect     int    `json:"windDirect"`
	Temperture     int    `json:"temperture"`
	Humidity       int    `json:"humidity"`
}

type DroneStatusResp struct {
	Code int `json:"code"`
}

// 飞行状态机
type FlightPhase int

const (
	Idle       FlightPhase = iota // 空闲状态
	TakeOff                       // 起飞
	Climbing                      // 爬升
	Cruising                      // 巡航
	Descending                    // 下降
	Land                          // 降落
)

// 判断是否到达目标点（5米内算到达）
func reachedTarget(lat, lon, targetLat, targetLon float64) bool {
	const threshold = 5.0 // 米
	const earthRadius = 6371000.0
	dLat := (targetLat - lat) * math.Pi / 180
	dLon := (targetLon - lon) * math.Pi / 180
	meanLat := (lat + targetLat) / 2 * math.Pi / 180
	dx := earthRadius * dLon * math.Cos(meanLat)
	dy := earthRadius * dLat
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance < threshold
}

// 判断当前位置是否有高楼
func isInBuildingArea(lat, lon float64) bool {
	// 示例：某区域有高楼
	return lat > 22.8008 && lat < 22.8009 && lon > 113.9531 && lon < 113.9532
}

func main() {
	// 命令行参数定义
	latPtr := flag.Float64("lat", 22.8007210, "初始纬度（如 22.8007210）")
	lonPtr := flag.Float64("lon", 113.9530990, "初始经度（如 113.9530990）")
	bearingPtr := flag.Int("bearing", 45.0, "初始飞行方向角度（0~360）")
	uasIdPtr := flag.String("id", "uas3", "无人机ID")

	flag.Parse()

	// 初始经纬度
	lat := *latPtr
	lon := *lonPtr
	bearing := *bearingPtr // 飞行方向（度），例如 45° 表示东北方向
	// 服务器 URL（请根据你的服务器地址修改）
	serverURL := "http://localhost:19999/api/drone/status"

	speed := 10.0 // 无人机速度（米/秒）
	interval := 1 // 每 1 秒计算一次位置
	orderID := *uasIdPtr
	flightCode := *uasIdPtr
	height := 0.0
	altitude := 0.0
	SOC := 100 // 电量百分比

	// 起飞和飞行高度
	cruiseHeight := 1000 // 巡航高度（米 * 10）
	flighrStatus := "TakeOff"

	// 目标点
	// targetLat := 22.8044729
	// targetLon := 113.9571690
	targetLat := 22.8019928
	targetLon := 113.9544786

	// 创建一个定时器，每秒执行一次
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()

	// 持续发送请求
	id := 0
	phase := Idle
	for range ticker.C {
		switch phase {
		case Idle:
			if altitude < 1 {
				flighrStatus = "TakeOff"
				fmt.Println("待机阶段")
				altitude += 1
			} else {
				phase = TakeOff
				fmt.Println("进入起飞阶段")
			}
		case TakeOff:
			if altitude < 5 {
				flighrStatus = "Inflight"
				altitude += 1
				fmt.Printf("起飞中，高度：%.1f 米\n", altitude)
			} else {
				phase = Climbing
				fmt.Println("进入爬升阶段")
			}
		case Climbing:
			if int(altitude*10) < cruiseHeight {
				altitude += 10
				fmt.Printf("爬升中，高度：%.1f 米\n", altitude)
			} else {
				phase = Cruising
				fmt.Println("进入巡航阶段")
			}
		case Cruising:
			// 移动无人机
			lat, lon = flight.GetNewLatLon(lat, lon, speed, bearing, interval)
			SOC -= 1 // 每秒消耗1%的电量
			if isInBuildingArea(lat, lon) && altitude < 120 {
				altitude = 120
			} else if !isInBuildingArea(lat, lon) && altitude > 100 {
				altitude = 100
			}
			// 检查降落条件
			if reachedTarget(lat, lon, targetLat, targetLon) || SOC < 20 {
				phase = Descending
				fmt.Println("进入下降阶段")
			}
		case Descending:
			if altitude > 5 {
				altitude -= 10
				if altitude < 5 {
					altitude = 5
				}
				fmt.Printf("下降中，高度：%.1f 米\n", altitude)
			} else {
				phase = Land
				fmt.Println("进入降落阶段")
			}
		case Land:
			if altitude > 1 {
				altitude -= 1
				fmt.Printf("降落中，高度：%.1f 米\n", altitude)
				if altitude == 1 {
					altitude = 0 // 模拟降落到地面
					flighrStatus = "Land"
				}
			} else {
				fmt.Println("无人机已降落，模拟结束。")
				return
			}
		}

		// 构造要发送的数据
		droneStatusReq := DroneStatusReq{
			OrderID:        orderID,
			FlightCode:     flightCode,
			Sn:             flightCode,
			FlightStatus:   flighrStatus,
			ManufacturerID: "112233",
			UasID:          "Uas-default",
			TimeStamp:      time.Now().Format("20060102150405"),
			UasModel:       "Uas-default",
			Coordinate:     1,
			Longitude:      int64(lon * 10000000), // 转换为整数
			Latitude:       int64(lat * 10000000), // 转换为整数
			HeightType:     1,
			Height:         int(height * 10),
			Altitude:       int(altitude * 10),
			VS:             int(speed * 10),
			GS:             0,
			Course:         int(bearing * 10),
			SOC:            SOC,
			RM:             10,
			WindSpeed:      10,
			WindDirect:     90,
			Temperture:     30,
			Humidity:       50,
		}

		jsonData, err := json.Marshal(droneStatusReq)
		if err != nil {
			break
		}

		// 发送 POST 请求
		resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			log.Println("POST 请求失败:", err)
			continue
		}

		// 读取并打印服务器响应
		fmt.Printf("Step %d: 纬度: %.7f, 经度: %.7f, 高度: %.1f, 电量: %d%%, 轨迹点类别: %s\n", id, lat, lon, altitude, SOC, flighrStatus)

		resp.Body.Close()

		id++
		// if id%30 == 0 {
		// 	record = time.Now().Format("20060102150405") + fmt.Sprintf("%03d", id)
		// 	bearing += 180.0
		// 	if bearing >= 360.0 {
		// 		bearing -= 360.0
		// 	}
		// }
	}
}
