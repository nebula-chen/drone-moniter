package handler

import (
	"net/http"

	"drone-api/internal/logic"
	"drone-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleFlightStatsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewHandleFlightStatsLogic(r.Context(), svcCtx)
		resp, err := l.HandleFlightStats()
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
