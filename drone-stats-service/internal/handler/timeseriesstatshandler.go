package handler

import (
	"net/http"

	"drone-stats-service/internal/logic"
	"drone-stats-service/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TimeSeriesStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewTimeSeriesStatsLogic(r.Context(), svcCtx)
		resp, err := l.TimeSeriesStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
