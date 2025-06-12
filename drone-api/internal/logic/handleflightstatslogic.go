package logic

import (
	"context"

	"drone-api/internal/svc"
	"drone-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleFlightStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleFlightStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFlightStatsLogic {
	return &HandleFlightStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleFlightStatsLogic) HandleFlightStats() (resp *types.FlightStatsResp, err error) {
	// 1. 查询统计数据
	totalFlights, totalDistance, err := l.svcCtx.Dao.HandleFlightStats()
	if err != nil {
		l.Logger.Errorf("查询飞行统计失败: %v", err)
		return nil, err
	}

	// 2. 返回响应
	return &types.FlightStatsResp{
		TotalFlights:  totalFlights,
		TotalDistance: totalDistance,
	}, nil
}
