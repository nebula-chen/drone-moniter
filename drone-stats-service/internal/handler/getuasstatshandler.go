package handler

import (
	"net/http"

	"drone-stats-service/internal/logic"
	"drone-stats-service/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetUasStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewGetUasStatsLogic(r.Context(), svcCtx)
		resp, err := l.GetUasStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
