package logic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"autonomous-vehicle/internal/dao"
	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ExportVehicleRecordsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewExportVehicleRecordsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ExportVehicleRecordsLogic {
	return &ExportVehicleRecordsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ExportVehicleRecords 查询并导出数据，返回临时文件路径
func (l *ExportVehicleRecordsLogic) ExportVehicleRecords(req *types.ExportVehicleRecordsReq) (string, error) {
	start, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return "", fmt.Errorf("开始时间格式错误: %w", err)
	}
	end, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return "", fmt.Errorf("结束时间格式错误: %w", err)
	}
	start = start.UTC()
	end = end.UTC()

	records, err := l.svcCtx.Dao.QueryVehicleData(start, end)
	if err != nil {
		return "", err
	}

	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("vehicle_records_%d.xlsx", time.Now().Unix()))
	if err := dao.ExportVehicleRecordsToExcel(records, tmpFile); err != nil {
		return "", err
	}
	return tmpFile, nil
}
