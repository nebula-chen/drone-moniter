package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SOCUsageStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSOCUsageStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SOCUsageStatsLogic {
	return &SOCUsageStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SOCUsageStatsLogic) SOCUsageStats() (resp *types.SOCUsageStatsResp, err error) {
	yearStats, monthStats, dayStats, err := l.svcCtx.MySQLDao.GetSOCUsageStats()
	if err != nil {
		return nil, err
	}
	resp = &types.SOCUsageStatsResp{}
	for _, y := range yearStats {
		resp.YearStats = append(resp.YearStats, types.SOCUsage{
			Date:  y["date"].(string),
			Usage: y["total"].(float64),
		})
	}
	for _, m := range monthStats {
		resp.MonthStats = append(resp.MonthStats, types.SOCUsage{
			Date:  m["date"].(string),
			Usage: m["total"].(float64),
		})
	}
	for _, d := range dayStats {
		resp.DayStats = append(resp.DayStats, types.SOCUsage{
			Date:  d["date"].(string),
			Usage: d["total"].(float64),
		})
	}
	return
}
