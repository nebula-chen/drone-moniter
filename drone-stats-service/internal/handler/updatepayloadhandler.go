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

		// 查询uasID和时间范围内的飞行记录
		records, err := svcCtx.MySQLDao.QueryFlightRecords("", req.UasID, req.StartTime, req.EndTime)
		if err != nil {
			httpx.OkJson(w, types.UpdatePayloadResp{
				Code:     1,
				ErrorMsg: "查询飞行记录失败: " + err.Error(),
			})
			return
		}
		if len(records) == 0 {
			httpx.OkJson(w, types.UpdatePayloadResp{
				Code:     1,
				ErrorMsg: "未找到对应飞行记录",
			})
			return
		}

		// 遍历所有查到的OrderID，批量更新
		for _, rec := range records {
			orderID, ok := rec["OrderID"].(string)
			if !ok {
				continue
			}
			err := svcCtx.MySQLDao.UpdateFlightPayload(orderID, req.Payload)
			if err != nil {
				httpx.OkJson(w, types.UpdatePayloadResp{
					Code:     1,
					ErrorMsg: "更新失败: " + err.Error(),
				})
				return
			}
		}

		httpx.OkJson(w, types.UpdatePayloadResp{
			Code: 0,
		})
	}
}
