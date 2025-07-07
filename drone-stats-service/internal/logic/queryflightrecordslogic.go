package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryFlightRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryFlightRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryFlightRecordsLogic {
	return &QueryFlightRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryFlightRecordsLogic) QueryFlightRecords(req *types.FlightRecordReq) (resp *types.FlightRecordsResponse, err error) {
	// 字段名已变更：OrderID, StartTime, EndTime
	records, err := l.svcCtx.MySQLDao.QueryFlightRecords(req.OrderID, req.UasID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	resp = &types.FlightRecordsResponse{}
	for _, r := range records {
		resp.Flightrecords = append(resp.Flightrecords, types.FlightRecord{
			ID:          r["id"].(int),
			OrderID:     r["OrderID"].(string),
			UasID:       r["uasID"].(string), // 新增
			StartTime:   r["start_time"].(string),
			EndTime:     r["end_time"].(string),
			StartLat:    r["start_lat"].(int64),
			StartLng:    r["start_lng"].(int64),
			EndLat:      r["end_lat"].(int64),
			EndLng:      r["end_lng"].(int64),
			Distance:    r["distance"].(float64),
			BatteryUsed: r["battery_used"].(int),
			CreatedAt:   r["created_at"].(string),
			Payload:     r["payload"].(int),
		})
	}
	return resp, nil
}
