package handler

import (
	"net/http"

	"drone-stats-service/internal/logic"
	"drone-stats-service/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func SOCUsageStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		l := logic.NewSOCUsageStatsLogic(r.Context(), svcCtx)
		resp, err := l.SOCUsageStats(mode)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
