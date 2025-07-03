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

	startRM := getInt64(startPoint, "RM")
	endRM := getInt64(endPoint, "RM")

	// 计算距离
	var totalDistance float64
	for i := 1; i < len(flightPoints); i++ {
		lat1 := float64(getInt64(flightPoints[i-1], "latitude")) / 1e7
		lng1 := float64(getInt64(flightPoints[i-1], "longitude")) / 1e7
		lat2 := float64(getInt64(flightPoints[i], "latitude")) / 1e7
		lng2 := float64(getInt64(flightPoints[i], "longitude")) / 1e7
		totalDistance += haversine(lat1, lng1, lat2, lng2)
	}

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
		BatteryUsed: int(startRM - endRM),
		Payload:     getFloat64(endPoint, "payload"),
	}

	// 新增：插入前判断是否已存在
	exists, err := l.svcCtx.MySQLDao.FlightRecordExists(fr.OrderID, fr.StartTime, fr.EndTime)
	if err != nil {
		return nil, err
	}
	if exists {
		l.Logger.Infof("该飞行架次已存在: uav_id=%s, start=%v, end=%v", fr.OrderID, fr.StartTime, fr.EndTime)
		return &types.TrackResponse{}, nil
	}

	orderID, err := l.svcCtx.MySQLDao.SaveFlightRecordAndGetOrderID(fr)
	if err != nil {
		return nil, err
	}

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

	// 返回整理后的飞行轨迹
	resp = &types.TrackResponse{}
	for _, r := range flightPoints {
		resp.Track = append(resp.Track, types.TrackPoints{
			OrderID:      req.OrderID,
			FlightStatus: getString(r, "flightStatus"),
			TimeStamp:    r["_time"].(time.Time).Format("20060102150405"), // 按API要求格式
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
			WindSpeed:    int(getInt64(r, "windSpeed")),
			WindDirect:   int(getInt64(r, "windDirect")),
			Temperture:   int(getInt64(r, "temperture")),
			Humidity:     int(getInt64(r, "humidity")),
		})
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
