package handler

import (
	"encoding/csv"
	"net/http"
	"strconv"

	"drone-stats-service/internal/svc"
)

func ExportFlightRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orderID := r.URL.Query().Get("OrderID")
		uasID := r.URL.Query().Get("uasID") // 新增uasID参数
		startTime := r.URL.Query().Get("startTime")
		endTime := r.URL.Query().Get("endTime")
		records, err := svcCtx.MySQLDao.QueryFlightRecords(orderID, uasID, startTime, endTime)
		if err != nil {
			http.Error(w, "查询失败", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment;filename=flight_records.csv")
		writer := csv.NewWriter(w)
		writer.Write([]string{"ID", "架次编号", "无人机编号", "起飞时间", "降落时间", "起飞纬度", "起飞经度", "降落纬度", "降落经度", "飞行距离", "电池使用量", "创建时间"})
		for _, r := range records {
			writer.Write([]string{
				strconv.Itoa(r["id"].(int)),
				r["OrderID"].(string),
				r["uasID"].(string), // 新增uasID
				r["start_time"].(string),
				r["end_time"].(string),
				strconv.FormatInt(r["start_lat"].(int64), 10),
				strconv.FormatInt(r["start_lng"].(int64), 10),
				strconv.FormatInt(r["end_lat"].(int64), 10),
				strconv.FormatInt(r["end_lng"].(int64), 10),
				strconv.FormatFloat(r["distance"].(float64), 'f', 2, 64),
				strconv.Itoa(r["battery_used"].(int)),
				r["created_at"].(string),
			})
		}
		writer.Flush()
	}
}
