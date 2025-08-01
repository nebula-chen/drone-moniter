package handler

import (
	"net/http"

	"autonomous-vehicle/internal/logic"
	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleGetVehicleInfoHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetVehicleInfoReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewHandleGetVehicleInfoLogic(r.Context(), svcCtx)
		resp, err := l.HandleGetVehicleInfo(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
