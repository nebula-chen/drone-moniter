package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AvgStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAvgStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AvgStatsLogic {
	return &AvgStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AvgStatsLogic) AvgStats() (resp *types.AvgStatsResp, err error) {
	avgTime, avgBattery, avgGS, err := l.svcCtx.MySQLDao.GetAvgStats()
	if err != nil {
		return nil, err
	}
	return &types.AvgStatsResp{
		AvgFlightTime:  avgTime,
		AvgBatteryUsed: avgBattery,
		AvgGS:          avgGS,
	}, nil
}
