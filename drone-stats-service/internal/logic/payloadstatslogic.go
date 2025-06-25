package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PayloadStatsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPayloadStatsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PayloadStatsLogic {
	return &PayloadStatsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PayloadStatsLogic) PayloadStats() (resp *types.PayloadStatsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
