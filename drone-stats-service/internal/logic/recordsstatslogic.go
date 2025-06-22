package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RecordsStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRecordsStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RecordsStatsLogic {
	return &RecordsStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RecordsStatsLogic) RecordsStats() (resp *types.RecordsStatsResp, err error) {
	totalCount, totalDistance, totalTime, err := l.svcCtx.MySQLDao.GetFlightStats()
	if err != nil {
		return nil, err
	}
	return &types.RecordsStatsResp{
		TotalCount:    totalCount,
		TotalDistance: totalDistance,
		TotalTime:     totalTime,
	}, nil
}
