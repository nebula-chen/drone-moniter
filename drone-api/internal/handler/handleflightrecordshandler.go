package handler

import (
	"net/http"

	"drone-api/internal/logic"
	"drone-api/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func HandleFlightRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 直接解析参数
		page := r.URL.Query().Get("page")
		pageSize := r.URL.Query().Get("pageSize")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		uavId := r.URL.Query().Get("uavId")

		l := logic.NewHandleFlightRecordsLogic(r.Context(), svcCtx)
		resp, err := l.HandleFlightRecords(page, pageSize, start, end, uavId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
