package handler

import (
	"net/http"

	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleOnlineCountHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		count := 0
		svcCtx.OnlineDrones.Range(func(_, _ interface{}) bool {
			count++
			return true
		})
		resp := types.OnlineCountResp{Count: count}
		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
