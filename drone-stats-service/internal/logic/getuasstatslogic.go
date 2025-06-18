package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUasStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUasStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUasStatsLogic {
	return &GetUasStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUasStatsLogic) GetUasStats() (resp *types.UasStatsResp, err error) {
	total, err := l.svcCtx.MySQLDao.CountTotalUas()
	if err != nil {
		return nil, err
	}
	online, err := l.svcCtx.MySQLDao.CountOnlineUas()
	if err != nil {
		return nil, err
	}
	resp = &types.UasStatsResp{
		Total:  total,
		Online: online,
	}
	return
}
