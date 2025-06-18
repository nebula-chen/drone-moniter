package handler

import (
	"net/http"

	"drone-stats-service/internal/logic"
	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetFlightRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.FlightRecordReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewGetFlightRecordsLogic(r.Context(), svcCtx)
		resp, err := l.GetFlightRecords(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
