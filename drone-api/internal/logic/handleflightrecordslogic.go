package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"drone-api/internal/svc"
	"drone-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type HandleFlightRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewHandleFlightRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *HandleFlightRecordsLogic {
	return &HandleFlightRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *HandleFlightRecordsLogic) HandleFlightRecords(
	pageStr, pageSizeStr, start, end, uavId string,
) (resp *types.FlightRecordListResp, err error) {
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 20
	}
	// 转换时间
	var startUnix, endUnix int64
	if start != "" {
		t, _ := time.Parse("2006-01-02", start)
		startUnix = t.Unix()
	}
	if end != "" {
		t, _ := time.Parse("2006-01-02", end)
		endUnix = t.Add(24 * time.Hour).Unix() // 包含当天
	}

	records, total, totalUav, totalFlights, err := l.svcCtx.Dao.QueryFlightRecords(page, pageSize, startUnix, endUnix, uavId)
	if err != nil {
		l.Logger.Errorf("查询飞行记录失败: %v", err)
		fmt.Printf("查询飞行记录失败: %+v\n", err) // 新增，直接输出到控制台
		return nil, err
	}
	resp = &types.FlightRecordListResp{
		Total:        total,
		Records:      records,
		TotalUav:     totalUav,
		TotalFlights: totalFlights,
	}
	return resp, nil
}
