package handler

import (
	"net/http"

	"drone-api/internal/logic"
	"drone-api/internal/svc"
	"drone-api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleDroneStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.DroneStatusReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewHandleDroneStatusLogic(r.Context(), svcCtx)
		resp, err := l.HandleDroneStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
