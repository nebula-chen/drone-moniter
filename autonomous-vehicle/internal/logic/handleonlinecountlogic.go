package logic

import (
	"context"

	"autonomous-vehicle/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleOnlineCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleOnlineCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleOnlineCountLogic {
	return &HandleOnlineCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleOnlineCountLogic) HandleOnlineCount() error {
	// todo: add your logic here and delete this line

	return nil
}
