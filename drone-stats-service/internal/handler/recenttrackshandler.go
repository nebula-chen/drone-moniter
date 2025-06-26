package handler

import (
	"net/http"
	"strconv"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func RecentTracksHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 支持 ?id=xxx 查询指定轨迹，否则查最近n条
		idStr := r.URL.Query().Get("id")
		var recordIds []int

		if idStr != "" {
			if id, err := strconv.Atoi(idStr); err == nil {
				recordIds = []int{id}
			} else {
				httpx.Error(w, err)
				return
			}
		} else {
			n := 3
			if val := r.URL.Query().Get("n"); val != "" {
				if num, err := strconv.Atoi(val); err == nil && num > 0 {
					n = num
				}
			}
			// 查询最近n条飞行记录
			records, err := svcCtx.MySQLDao.QueryFlightRecords("", "", "")
			if err != nil {
				httpx.Error(w, err)
				return
			}
			for i := 0; i < n && i < len(records); i++ {
				if id, ok := records[i]["id"].(int); ok {
					recordIds = append(recordIds, id)
				}
			}
		}

		var allPoints []types.TrackPoints
		for _, rid := range recordIds {
			points, err := svcCtx.MySQLDao.GetTrackPointsByRecordId(rid)
			if err != nil {
				httpx.Error(w, err)
				return
			}
			// 查询主表获得FlightCode
			var flightCode string
			row := svcCtx.MySQLDao.DB.QueryRow("SELECT uav_id FROM flight_records WHERE id=?", rid)
			row.Scan(&flightCode)
			for _, pt := range points {
				allPoints = append(allPoints, types.TrackPoints{
					FlightCode:   flightCode,
					FlightStatus: pt["flightStatus"].(string),
					TimeStamp:    pt["timeStamp"].(string),
					Longitude:    pt["longitude"].(int64),
					Latitude:     pt["latitude"].(int64),
					Altitude:     pt["altitude"].(int),
					SOC:          pt["SOC"].(int),
					GS:           pt["GS"].(int),
				})
			}
		}
		resp := types.TrackResponse{
			Track: allPoints,
		}
		httpx.OkJson(w, resp)
	}
}
