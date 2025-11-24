package handler

import (
	"encoding/json"
	"net/http"

	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"
)

// 创建导出任务（异步）
// POST /record/exportAsync
// body: { startTime, endTime, OrderID, uasID, target }
func CreateExportTaskHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		// map -> types
		req := types.FlightRecordReq{}
		if v, ok := raw["startTime"].(string); ok {
			req.StartTime = v
		}
		if v, ok := raw["endTime"].(string); ok {
			req.EndTime = v
		}
		if v, ok := raw["OrderID"].(string); ok {
			req.OrderID = v
		}
		if v, ok := raw["uasID"].(string); ok {
			req.UasID = v
		}
		target := "records"
		if v, ok := raw["target"].(string); ok && v != "" {
			target = v
		}
		format := "xlsx"
		if v, ok := raw["format"].(string); ok && v != "" {
			format = v
		}
		if svcCtx.TaskManager == nil {
			http.Error(w, "TaskManager 未启用", http.StatusInternalServerError)
			return
		}
		id, err := svcCtx.TaskManager.CreateTask(target, req.OrderID, req.UasID, req.StartTime, req.EndTime, format)
		if err != nil {
			http.Error(w, "创建任务失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// 返回完整下载/状态 URL（TaskManager 负责拼接 baseURL）
		statusUrl := "/record/exportStatus?id=" + id
		downloadUrl := "/record/exportDownload?id=" + id
		if svcCtx.TaskManager != nil {
			statusUrl = svcCtx.TaskManager.StatusURL(id)
			downloadUrl = svcCtx.TaskManager.DownloadURL(id)
		}
		resp := map[string]string{
			"taskId":      id,
			"statusUrl":   statusUrl,
			"downloadUrl": downloadUrl,
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// 查询导出任务状态
// GET /record/exportStatus?id=xxx
func ExportStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if svcCtx.TaskManager == nil {
			http.Error(w, "TaskManager 未启用", http.StatusInternalServerError)
			return
		}
		t, ok := svcCtx.TaskManager.GetTask(id)
		if !ok {
			http.Error(w, "任务未找到", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(t)
	}
}

// 下载导出结果
// GET /record/exportDownload?id=xxx
func ExportDownloadHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		if svcCtx.TaskManager == nil {
			http.Error(w, "TaskManager 未启用", http.StatusInternalServerError)
			return
		}
		t, ok := svcCtx.TaskManager.GetTask(id)
		if !ok {
			http.Error(w, "任务未找到", http.StatusNotFound)
			return
		}
		if t.Status != "done" {
			http.Error(w, "任务未完成", http.StatusBadRequest)
			return
		}
		// 返回文件
		http.ServeFile(w, r, t.ResultFile)
	}
}
