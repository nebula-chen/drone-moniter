package handler

import (
	"net/http"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func UpdatePayloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.UpdatePayloadReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, err)
			return
		}
		err := svcCtx.MySQLDao.UpdateFlightPayload(req.OrderID, req.Payload)
		if err != nil {
			httpx.OkJson(w, types.UpdatePayloadResp{
				Code:     1,
				ErrorMsg: "更新失败: " + err.Error(),
			})
			return
		}
		httpx.OkJson(w, types.UpdatePayloadResp{
			Code: 0,
		})
	}
}
