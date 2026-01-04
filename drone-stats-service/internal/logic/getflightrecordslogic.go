package logic

import (
	"context"
	"drone-stats-service/internal/model"
	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"
	"fmt"
	"math"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFlightRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFlightRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFlightRecordsLogic {
	return &GetFlightRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFlightRecordsLogic) GetFlightRecords(req *types.FlightRecordReq) (resp *types.TrackResponse, err error) {
	start, _ := time.Parse(time.RFC3339, req.StartTime)
	end, _ := time.Parse(time.RFC3339, req.EndTime)
	// 转为UTC0
	start = start.UTC()
	end = end.UTC()

	records, err := l.svcCtx.InfluxDao.QueryFlightRecords(req.OrderID, start, end)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return &types.TrackResponse{}, nil
	}

	// 找到TakeOff和Land点，提取本次飞行架次的所有点
	var (
		flightPoints []map[string]interface{}
		inFlight     bool
	)
	for _, r := range records {
		status, ok := r["flightStatus"].(string)
		if !ok {
			continue // 只处理flightStatus字段
		}
		if status == "TakeOff" {
			inFlight = true
			flightPoints = []map[string]interface{}{r}
		} else if status == "Inflight" && inFlight {
			flightPoints = append(flightPoints, r)
		} else if status == "Land" && inFlight {
			flightPoints = append(flightPoints, r)
			break // 一次飞行架次结束
		}
	}

	getInt64 := func(m map[string]interface{}, key string) int64 {
		if v, ok := m[key]; ok && v != nil {
			if val, ok := v.(int64); ok {
				return val
			}
			if val, ok := v.(float64); ok {
				return int64(val)
			}
		}
		return 0
	} // 获取int64类型的值，支持float64转换
	getString := func(m map[string]interface{}, key string) string {
		if v, ok := m[key]; ok && v != nil {
			if val, ok := v.(string); ok {
				return val
			}
		}
		return ""
	} // 获取string类型的值
	getFloat64 := func(m map[string]interface{}, key string) float64 {
		if v, ok := m[key]; ok && v != nil {
			switch val := v.(type) {
			case float64:
				return val
			case int64:
				return float64(val)
			}
		}
		return 0
	} // 获取float64类型的值，支持int64转换

	if len(flightPoints) < 2 ||
		getString(flightPoints[0], "flightStatus") != "TakeOff" ||
		getString(flightPoints[len(flightPoints)-1], "flightStatus") != "Land" {
		// return &types.TrackResponse{}, nil // 无有效飞行架次

		// 仅打印到终端，不存入MySQL
		for i, r := range records {
			l.Logger.Infof("无效架次: index=%d, record=%+v 结束", i, r)
		}
		return &types.TrackResponse{}, nil
	}

	startPoint := flightPoints[0]
	endPoint := flightPoints[len(flightPoints)-1]

	energyUsed := calcEnergyKWh(flightPoints)

	// 计算距离：
	// 1) 水平距离使用起点与终点的球面距离（基于经纬度）
	// 2) 垂直距离为全过程上下移动距离（每段海拔变化的绝对值之和）
	// 最终距离 = 水平距离 + 垂直移动距离（按用户要求将两者相加）
	var totalDistance float64
	// 起终点水平球面距离
	startLat := float64(getInt64(flightPoints[0], "latitude")) / 1e7
	startLng := float64(getInt64(flightPoints[0], "longitude")) / 1e7
	endLat := float64(getInt64(flightPoints[len(flightPoints)-1], "latitude")) / 1e7
	endLng := float64(getInt64(flightPoints[len(flightPoints)-1], "longitude")) / 1e7
	horizontal := haversine(startLat, startLng, endLat, endLng)

	// 全过程垂直移动距离（海拔单位按原代码除以10）
	var verticalMovement float64
	for i := 1; i < len(flightPoints); i++ {
		altPrev := float64(getInt64(flightPoints[i-1], "altitude")) / 10
		altCurr := float64(getInt64(flightPoints[i], "altitude")) / 10
		verticalMovement += math.Abs(altCurr - altPrev)
	}

	totalDistance = horizontal + verticalMovement

	// 存储到flight_records主表，注意经纬度/高度转换
	fr := model.FlightRecord{
		OrderID:     req.OrderID,
		UasID:       getString(startPoint, "uasID"), // 新增，确保从influx数据中获取uasID
		StartTime:   startPoint["_time"].(time.Time),
		EndTime:     endPoint["_time"].(time.Time),
		StartLat:    getInt64(startPoint, "latitude"),
		StartLng:    getInt64(startPoint, "longitude"),
		EndLat:      getInt64(endPoint, "latitude"),
		EndLng:      getInt64(endPoint, "longitude"),
		Distance:    totalDistance,
		BatteryUsed: energyUsed,
		Payload:     getFloat64(endPoint, "payload"),
	}

	// 新增：插入前判断是否已存在
	exists, err := l.svcCtx.MySQLDao.FlightRecordExists(fr.OrderID, fr.StartTime, fr.EndTime)
	if err != nil {
		return nil, err
	}
	if exists {
		l.Logger.Infof("该飞行架次已存在: uav_id=%s, start=%v, end=%v", fr.OrderID, fr.StartTime, fr.EndTime)
		// 如果主表存在，但轨迹点为空，则尝试回填轨迹点
		pts, err := l.svcCtx.MySQLDao.GetTrackPointsByRecordId(fr.OrderID)
		if err != nil {
			return nil, err
		}
		if len(pts) == 0 {
			// 构造并保存轨迹点（使用字符串 OrderID）
			var trackPoints []model.FlightTrackPoint
			for _, r := range flightPoints {
				point := model.FlightTrackPoint{
					OrderID:      fr.OrderID,
					FlightStatus: getString(r, "flightStatus"),
					TimeStamp:    r["_time"].(time.Time),
					Longitude:    getInt64(r, "longitude"),
					Latitude:     getInt64(r, "latitude"),
					HeightType:   int(getInt64(r, "heightType")),
					Height:       int(getInt64(r, "height")),
					Altitude:     int(getInt64(r, "altitude")),
					VS:           int(getInt64(r, "VS")),
					GS:           int(getInt64(r, "GS")),
					Course:       int(getInt64(r, "course")),
					SOC:          int(getInt64(r, "SOC")),
					RM:           int(getInt64(r, "RM")),
					Voltage:      int(getInt64(r, "voltage")),
					Current:      int(getInt64(r, "current")),
					WindSpeed:    int(getInt64(r, "windSpeed")),
					WindDirect:   int(getInt64(r, "windDirect")),
					Temperture:   int(getInt64(r, "temperture")),
					Humidity:     int(getInt64(r, "humidity")),
				}
				trackPoints = append(trackPoints, point)
			}
			if err := l.svcCtx.MySQLDao.SaveTrackPoints(trackPoints); err != nil {
				fmt.Println("回填轨迹点失败:", err)
			} else {
				l.Logger.Infof("已为存在的飞行架次回填轨迹点: %s", fr.OrderID)
			}
		}
		return &types.TrackResponse{}, nil
	}

	orderID, err := l.svcCtx.MySQLDao.SaveFlightRecordAndGetOrderID(fr)
	if err != nil {
		return nil, err
	} //else {
	// 	l.Logger.Infof("当前架次: %s, 最终距离=%.2f, 水平距离=%.2f, 垂直移动距离=%.2f", fr.OrderID, totalDistance, horizontal, verticalMovement)
	// }

	// 批量构造轨迹点，结构与flight_record.go同步
	var trackPoints []model.FlightTrackPoint
	for _, r := range flightPoints {
		point := model.FlightTrackPoint{
			OrderID:      orderID,
			FlightStatus: getString(r, "flightStatus"),
			TimeStamp:    r["_time"].(time.Time),
			Longitude:    getInt64(r, "longitude"),
			Latitude:     getInt64(r, "latitude"),
			HeightType:   int(getInt64(r, "heightType")),
			Height:       int(getInt64(r, "height")),
			Altitude:     int(getInt64(r, "altitude")),
			VS:           int(getInt64(r, "VS")),
			GS:           int(getInt64(r, "GS")),
			Course:       int(getInt64(r, "course")),
			SOC:          int(getInt64(r, "SOC")),
			RM:           int(getInt64(r, "RM")),
			Voltage:      int(getInt64(r, "voltage")),
			Current:      int(getInt64(r, "current")),
			WindSpeed:    int(getInt64(r, "windSpeed")),
			WindDirect:   int(getInt64(r, "windDirect")),
			Temperture:   int(getInt64(r, "temperture")),
			Humidity:     int(getInt64(r, "humidity")),
		}
		trackPoints = append(trackPoints, point)
	}

	// 一次性批量插入
	err = l.svcCtx.MySQLDao.SaveTrackPoints(trackPoints)
	if err != nil {
		fmt.Println("批量插入轨迹点失败:", err)
	}

	return
}

// 计算两点间球面距离（单位：米）
func haversine(lat1, lng1, lat2, lng2 float64) float64 {
	const R = 6371000 // 地球半径，单位米
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

// 计算所有轨迹点消耗的电能（单位：kWh）
func calcEnergyKWh(points []map[string]interface{}) float64 {
	var totalEnergyWh float64
	for _, p := range points {
		voltage := 0.0
		current := 0.0
		// 支持多种类型
		switch v := p["voltage"].(type) {
		case int:
			voltage = float64(v)
		case int64:
			voltage = float64(v)
		case float64:
			voltage = v
		}
		switch c := p["current"].(type) {
		case int:
			current = float64(c)
		case int64:
			current = float64(c)
		case float64:
			current = c
		}
		voltageV := voltage / 1000.0
		currentA := current / 1000.0
		totalEnergyWh += voltageV * currentA / 3600.0
	}
	return totalEnergyWh / 1000.0
}
