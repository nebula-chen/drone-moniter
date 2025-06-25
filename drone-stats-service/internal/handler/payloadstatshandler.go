package handler

import (
	"net/http"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func PayloadStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yearStats, monthStats, dayStats, err := svcCtx.MySQLDao.GetPayloadStats()
		if err != nil {
			httpx.Error(w, err)
			return
		}
		resp := types.PayloadStatsResp{
			YearStats:  convertPayloadStats(yearStats),
			MonthStats: convertPayloadStats(monthStats),
			DayStats:   convertPayloadStats(dayStats),
		}
		httpx.OkJson(w, resp)
	}
}

func convertPayloadStats(stats []map[string]interface{}) []types.PayloadStats {
	var res []types.PayloadStats
	for _, s := range stats {
		res = append(res, types.PayloadStats{
			Date:    s["date"].(string),
			Payload: s["payload"].(float64),
		})
	}
	return res
}
