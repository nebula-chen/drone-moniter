package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type TimeSeriesStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewTimeSeriesStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *TimeSeriesStatsLogic {
	return &TimeSeriesStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *TimeSeriesStatsLogic) TimeSeriesStats() (resp *types.TimeSeriesStatsResp, err error) {
	yearStats, monthStats, dayStats, err := l.svcCtx.MySQLDao.GetFlightRecordsStats()
	if err != nil {
		return nil, err
	}
	resp = &types.TimeSeriesStatsResp{}
	for _, y := range yearStats {
		resp.YearStats = append(resp.YearStats, types.DateCount{
			Date:  y["date"].(string),
			Count: y["count"].(int),
		})
	}
	for _, m := range monthStats {
		resp.MonthStats = append(resp.MonthStats, types.DateCount{
			Date:  m["date"].(string),
			Count: m["count"].(int),
		})
	}
	for _, d := range dayStats {
		resp.DayStats = append(resp.DayStats, types.DateCount{
			Date:  d["date"].(string),
			Count: d["count"].(int),
		})
	}
	return
}
