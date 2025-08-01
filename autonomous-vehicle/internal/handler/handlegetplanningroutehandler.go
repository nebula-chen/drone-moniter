package handler

import (
	"net/http"

	"autonomous-vehicle/internal/logic"
	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleGetPlanningRouteHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.GetPlanningRouteReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := logic.NewHandleGetPlanningRouteLogic(r.Context(), svcCtx)
		resp, err := l.HandleGetPlanningRoute(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
