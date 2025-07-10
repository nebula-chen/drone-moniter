package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"drone-stats-service/internal/dao"
	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"
)

func ExportFlightRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlightRecordReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		start, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			http.Error(w, "开始时间格式错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		end, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			http.Error(w, "结束时间格式错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		start = start.UTC()
		end = end.UTC()
		records, err := svcCtx.InfluxDao.GetFlightDate(start, end)
		if err != nil {
			http.Error(w, "查询失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// 临时文件路径
		tmpFile := filepath.Join(os.TempDir(), "flight_records.xlsx")
		// 导出为Excel
		err = dao.ExportFlightRecordsToExcel(records, tmpFile)
		if err != nil {
			http.Error(w, "导出失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpFile)

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=flight_records.xlsx")
		http.ServeFile(w, r, tmpFile)
	}
}
