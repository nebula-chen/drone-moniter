package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePayloadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePayloadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePayloadLogic {
	return &UpdatePayloadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePayloadLogic) UpdatePayload(req *types.UpdatePayloadReq) (resp *types.UpdatePayloadResp, err error) {
	// todo: add your logic here and delete this line

	return
}
