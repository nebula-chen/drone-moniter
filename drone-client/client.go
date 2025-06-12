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
	OrderID        string `json:"orderId"`
	FlightCode     string `json:"flightCode"`
	SN             string `json:"sn"`
	FlightStatus   string `json:"flightStatus"`
	ManufacturerID string `json:"manufacturerId"`
	UasId          string `json:"uasId"`
	TimeStamp      string `json:"timeStamp"`
	UasModel       string `json:"uasModel"`
	Coordinate     int    `json:"coordinate"`
	Longitude      int64  `json:"longitude"`
	Latitude       int64  `json:"latitude"`
	HeightType     int    `json:"heightType"`
	Height         int    `json:"height"`
	Altitude       int    `json:"altitude"`
	VS             int    `json:"vs"`
	GS             int    `json:"gs"`
	Course         int    `json:"course"`
	SOC            int    `json:"soc"`
	RM             int    `json:"rm"`
	WindSpeed      int    `json:"windSpeed"`
	WindDirect     int    `json:"windDirect"`
	Temperture     int    `json:"temperture"`
	Humidity       int    `json:"humidity"`
	RecordId       string `json:"recordId"`
	UavType        string `json:"uavType"`
	FlightTime     int    `json:"flightTime"`
	Payload        int    `json:"payload"`
}

type DroneStatusResp struct {
	Code int `json:"code"`
}

// 飞行状态机
type FlightPhase int

const (
	Idle FlightPhase = iota
	Takeoff
	Climbing
	Cruising
	Descending
	Land
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
	uasIdPtr := flag.String("id", "uav2", "无人机ID")
	payloadPtr := flag.Int("payload", 10, "载荷（单位：千克）")
	socPtr := flag.Int("soc", 80, "电池电量（百分比）")

	flag.Parse()

	lat := *latPtr
	lon := *lonPtr
	bearing := *bearingPtr // 飞行方向（度），例如 45° 表示东北方向
	payload := *payloadPtr
	soc := *socPtr
	// 服务器 URL（请根据你的服务器地址修改）
	serverURL := "http://localhost:19999/api/drone/status"

	// 初始经纬度
	speed := 10.0 // 无人机速度（米/秒）
	interval := 1 // 每 1 秒计算一次位置
	uavType := "typeA"
	uasId := *uasIdPtr
	flightTime := 0
	altitude := 160

	// 起飞和飞行高度
	cruiseHeight := 100
	height := 0

	// 目标点
	targetLat := 22.8044729
	targetLon := 113.9571690
	// targetLat := 22.8019928
	// targetLon := 113.9544786

	// 创建一个定时器，每秒执行一次
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()

	// 持续发送请求
	id := 0
	record := time.Now().Format("20060102150405") + fmt.Sprintf("%03d", id)
	phase := Idle
	for range ticker.C {
		switch phase {
		case Idle:
			if height == 0 && flightTime > 0 {
				phase = Takeoff
				fmt.Println("进入起飞阶段")
			}
		case Takeoff:
			if height < 5 {
				height += 1
				fmt.Printf("起飞中，高度：%d 米\n", height)
			} else {
				phase = Climbing
				fmt.Println("进入爬升阶段")
			}
		case Climbing:
			if height < cruiseHeight {
				height += 10
				fmt.Printf("爬升中，高度：%d 米\n", height)
			} else {
				phase = Cruising
				fmt.Println("进入巡航阶段")
			}
		case Cruising:
			// 移动无人机
			lat, lon = flight.GetNewLatLon(lat, lon, speed, bearing, interval)
			soc -= 1 // 每秒消耗1%的电量
			if isInBuildingArea(lat, lon) && height < 120 {
				height = 120
			} else if !isInBuildingArea(lat, lon) && height > 100 {
				height = 100
			}
			// 检查降落条件
			if reachedTarget(lat, lon, targetLat, targetLon) || soc < 20 {
				phase = Descending
				fmt.Println("进入下降阶段")
			}
		case Descending:
			if height > 5 {
				height -= 10
				if height < 5 {
					height = 5
				}
				fmt.Printf("下降中，高度：%d 米\n", height)
			} else {
				phase = Land
				fmt.Println("进入降落阶段")
			}
		case Land:
			if height > 1 {
				height -= 1
				fmt.Printf("降落中，高度：%d 米\n", height)
			} else {
				fmt.Println("无人机已降落，模拟结束。")
				return
			}
		}

		// 构造要发送的数据
		droneStatusReq := DroneStatusReq{
			OrderID:        "",
			FlightCode:     "",
			SN:             "",
			FlightStatus:   "",
			ManufacturerID: "",
			UasId:          uasId,
			TimeStamp:      time.Now().Format("2006-01-02-15-04-05"),
			UasModel:       "",
			Coordinate:     1,                     // 1 表示 WGS-84 坐标系
			Longitude:      int64(lon * 10000000), // 转换为整数
			Latitude:       int64(lat * 10000000), // 转换为整数
			HeightType:     1,
			Height:         height,
			Altitude:       altitude,
			VS:             0,
			GS:             0,
			Course:         bearing,
			SOC:            soc,
			RM:             0,
			WindSpeed:      0,
			WindDirect:     0,
			Temperture:     0,
			Humidity:       0,
			RecordId:       record,
			UavType:        uavType,
			FlightTime:     flightTime,
			Payload:        payload,
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
		fmt.Printf("Step %d: 纬度: %.7f, 经度: %.7f, 高度: %d, 电量: %d%%\n", id, lat, lon, height, soc)

		resp.Body.Close()

		id++
		flightTime += interval
		// if id%30 == 0 {
		// 	record = time.Now().Format("20060102150405") + fmt.Sprintf("%03d", id)
		// 	bearing += 180.0
		// 	if bearing >= 360.0 {
		// 		bearing -= 360.0
		// 	}
		// }
	}
}
