package logic

import (
	"context"
	"encoding/json"
	"math"
	"sync"
	"time"

	"drone-api/internal/svc"
	"drone-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleDroneStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

type FlightPhase int

const (
	Idle       FlightPhase = iota // 待机,未起飞
	Takeoff                       // 起飞
	Climbing                      // 爬升
	Cruising                      // 巡航
	Descending                    // 下降
	Land                          // 降落
)

type flightState struct {
	Phase        FlightPhase
	StartTime    time.Time
	StartLat     int64
	StartLng     int64
	StartBattery int
	LastLat      int64
	LastLng      int64
	LastHeight   int
	LastBattery  int
	LastTime     time.Time
}

var (
	flightStateMap = make(map[string]*flightState)
	flightStateMu  sync.Mutex
)

func NewHandleDroneStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDroneStatusLogic {
	return &HandleDroneStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func calcDistance(lat1, lng1, lat2, lng2 int64) int64 {
	// 使用haversine公式
	const R = 6371000 // 地球半径，单位为米
	lat1Rad := float64(lat1) * (math.Pi / 180.0)
	lat2Rad := float64(lat2) * (math.Pi / 180.0)
	deltaLat := (float64(lat2) - float64(lat1)) * (math.Pi / 180.0)
	deltaLng := (float64(lng2) - float64(lng1)) * (math.Pi / 180.0)
	a := (math.Sin(deltaLat/2) * math.Sin(deltaLat/2)) +
		(math.Cos(lat1Rad) * math.Cos(lat2Rad) * math.Sin(deltaLng/2) * math.Sin(deltaLng/2))
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return int64(R * c) // 返回距离，单位为米
}

func (l *HandleDroneStatusLogic) HandleDroneStatus(req *types.DroneStatusReq) (resp *types.DroneStatusResp, err error) {
	t, err := time.Parse("2006-01-02-15-04-05", req.TimeStamp)
	if err != nil {
		return nil, err
	}
	req.TimeStamp = t.Format(time.RFC3339)
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	l.svcCtx.WSHub.Broadcast <- []byte(data)

	// 1. 构建点
	point, err := l.svcCtx.Dao.BuildPoint(req)
	if err != nil {
		return nil, err
	}
	// 2. 写入点
	_ = l.svcCtx.Dao.AddPoint(point)

	// 3. 自动整理飞行记录
	flightStateMu.Lock()
	defer flightStateMu.Unlock()
	state, ok := flightStateMap[req.UasId]
	if !ok {
		state = &flightState{
			Phase: Idle,
		}
		flightStateMap[req.UasId] = state
	}

	// 状态机切换逻辑
	switch state.Phase {
	case Idle, Land:
		if req.Height > 1 && req.FlightTime > 0 {
			state.Phase = Takeoff
			state.StartTime = t
			state.StartLat = req.Latitude
			state.StartLng = req.Longitude
			state.StartBattery = req.SOC
		}
	case Takeoff:
		if req.Height > 5 {
			state.Phase = Climbing
		}
	case Climbing:
		if req.Height > 15 {
			state.Phase = Cruising
		}
	case Cruising:
		if req.Height < 100 {
			state.Phase = Descending
		}
	case Descending:
		if req.Height < 5 {
			state.Phase = Land
			logx.Info("写入flight_record") // 添加日志
			// 记录飞行日志
			endTime := t
			endLat := req.Latitude
			endLng := req.Longitude
			batteryUsed := state.StartBattery - req.SOC
			distance := calcDistance(state.StartLat, state.StartLng, endLat, endLng)
			record := &types.FlightRecordReq{
				StartTime:   state.StartTime.Format(time.RFC3339),
				EndTime:     endTime.Format(time.RFC3339),
				UasId:       req.UasId,
				Distance:    distance,
				BatteryUsed: batteryUsed,
				StartLat:    state.StartLat,
				StartLng:    state.StartLng,
				EndLat:      endLat,
				EndLng:      endLng,
			}
			err := l.svcCtx.Dao.AddFlightRecord(record)
			if err != nil {
				logx.Errorf("AddFlightRecord error: %v", err)
			}
		}
	}

	// 更新最近状态
	state.LastLat = req.Latitude
	state.LastLng = req.Longitude
	state.LastHeight = req.Height
	state.LastTime = t
	state.LastBattery = req.SOC

	resp = &types.DroneStatusResp{
		Code: "200",
	}
	return resp, nil
}
