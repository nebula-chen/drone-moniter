package logic

import (
	"context"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportFlightRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportFlightRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportFlightRecordsLogic {
	return &ExportFlightRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ExportFlightRecordsLogic) ExportFlightRecords(req *types.FlightRecordReq) (resp *types.FlightRecordsResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
