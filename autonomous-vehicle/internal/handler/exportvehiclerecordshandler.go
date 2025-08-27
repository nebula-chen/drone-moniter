package handler

import (
	"encoding/json"
	"net/http"
	"os"

	"autonomous-vehicle/internal/logic"
	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"
)

func ExportVehicleRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ExportVehicleRecordsReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
			return
		}

		lg := logic.NewExportVehicleRecordsLogic(r.Context(), svcCtx)
		filePath, err := lg.ExportVehicleRecords(&req)
		if err != nil {
			http.Error(w, "导出失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(filePath)

		w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		w.Header().Set("Content-Disposition", "attachment; filename=vehicle_records.xlsx")
		http.ServeFile(w, r, filePath)
	}
}
