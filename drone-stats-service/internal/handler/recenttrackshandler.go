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
		// 支持 ?orderID=xxx 查询指定轨迹，否则查最近n条
		orderID := r.URL.Query().Get("orderID")
		var orderIDs []string

		if orderID != "" {
			orderIDs = []string{orderID}
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
				if oid, ok := records[i]["OrderID"].(string); ok {
					orderIDs = append(orderIDs, oid)
				}
			}
		}

		var allPoints []types.TrackPoints
		for _, oid := range orderIDs {
			points, err := svcCtx.MySQLDao.GetTrackPointsByRecordId(oid)
			if err != nil {
				httpx.Error(w, err)
				return
			}
			for _, pt := range points {
				allPoints = append(allPoints, types.TrackPoints{
					OrderID:      pt["orderID"].(string),
					FlightStatus: pt["flightStatus"].(string),
					TimeStamp:    pt["timeStamp"].(string),
					Longitude:    pt["longitude"].(int64),
					Latitude:     pt["latitude"].(int64),
					HeightType:   pt["heightType"].(int),
					Height:       pt["height"].(int),
					Altitude:     pt["altitude"].(int),
					VS:           pt["VS"].(int),
					GS:           pt["GS"].(int),
					Course:       pt["course"].(int),
					SOC:          pt["SOC"].(int),
					RM:           pt["RM"].(int),
					WindSpeed:    pt["windSpeed"].(int),
					WindDirect:   pt["windDirect"].(int),
					Temperture:   pt["temperture"].(int),
					Humidity:     pt["humidity"].(int),
				})
			}
		}
		if allPoints == nil {
			allPoints = []types.TrackPoints{}
		}
		resp := types.TrackResponse{
			Track: allPoints,
		}
		httpx.OkJson(w, resp)
	}
}
