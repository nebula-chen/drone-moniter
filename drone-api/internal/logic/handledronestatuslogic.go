package logic

import (
	"context"
	"encoding/json"
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

func NewHandleDroneStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleDroneStatusLogic {
	return &HandleDroneStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleDroneStatusLogic) HandleDroneStatus(req *types.DroneStatusReq) (resp *types.DroneStatusResp, err error) {
	t, err := time.Parse("20060102150405", req.TimeStamp)
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

	resp = &types.DroneStatusResp{
		Code: "200",
	}
	return resp, nil
}
