package main

import (
	"bytes"
	"client/flight"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

type DroneStatusReq struct {
	RecordId   string  `json:"recordId"`
	UavType    string  `json:"uavType"`
	UavId      string  `json:"uavId"`
	TimeStamp  string  `json:"timeStamp"`
	FlightTime int     `json:"flightTime"`
	Longitude  int     `json:"longitude"`
	Latitude   int     `json:"latitude"`
	Altitude   float64 `json:"altitude"`
	Height     float64 `json:"height"`
	Course     float64 `json:"course"`
}

type DroneStatusResp struct {
	Code int `json:"code"`
}

func main() {
	// 命令行参数定义
	latPtr := flag.Float64("lat", 22.8007210, "初始纬度（如 22.8007210）")
	lonPtr := flag.Float64("lon", 113.9530990, "初始经度（如 113.9530990）")
	bearingPtr := flag.Float64("bearing", 45.0, "初始飞行方向角度（0~360）")
	uavIdPtr := flag.String("id", "uav1", "无人机ID")

	flag.Parse()

	lat := *latPtr
	lon := *lonPtr
	bearing := *bearingPtr // 飞行方向（度），例如 45° 表示东北方向
	// 服务器 URL（请根据你的服务器地址修改）
	serverURL := "http://localhost:19999/api/drone/status"

	// 初始经纬度
	speed := 10.0 // 无人机速度（米/秒）
	interval := 1 // 每 1 秒计算一次位置
	uavType := "typeA"
	uavId := *uavIdPtr
	flightTime := 0
	height := 150.0
	altitude := 160.0

	// 创建一个定时器，每秒执行一次
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 持续发送请求
	id := 0
	record := time.Now().Format("20060102150405") + fmt.Sprintf("%03d", id)
	for range ticker.C {
		// 构造要发送的数据
		droneStatusReq := DroneStatusReq{
			RecordId:   record,
			UavType:    uavType,
			UavId:      uavId,
			TimeStamp:  time.Now().Format("2006-01-02-15-04-05"),
			FlightTime: flightTime,
			Longitude:  int(lon * 10000000), // 转换为整数
			Latitude:   int(lat * 10000000), // 转换为整数
			Altitude:   altitude,
			Height:     height,
			Course:     bearing,
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
		fmt.Printf("Step %d: 纬度: %.7f, 经度: %.7f\n", id, lat, lon)

		lat, lon = flight.GetNewLatLon(lat, lon, speed, bearing, interval)
		resp.Body.Close()

		id++
		flightTime += interval
		if id%30 == 0 {
			record = time.Now().Format("20060102150405") + fmt.Sprintf("%03d", id)
			bearing += 180.0
			if bearing >= 360.0 {
				bearing -= 360.0
			}
		}
	}
}
