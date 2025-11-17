package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"drone-stats-service/internal/dao"
	"drone-stats-service/internal/export"
	"drone-stats-service/internal/svc"
	"drone-stats-service/internal/types"
)

// 支持导出三类：records（飞行主表 flight_records）、trajectory（轨迹表 flight_track_points）、both（两者一起，返回 zip）
// 导出优先使用 Influx（仅当时间跨度 <= 72 小时）导出 records；trajectory 始终从 MySQL 导出。
func ExportFlightRecordsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 读取 body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "读取请求体失败: "+err.Error(), http.StatusBadRequest)
			return
		}
		// 解两次：一份到 types.FlightRecordReq（核心字段），一份到 map 获取可选参数
		var req types.FlightRecordReq
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "请求参数错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		var raw map[string]interface{}
		_ = json.Unmarshal(body, &raw)
		target := "records"
		if v, ok := raw["target"].(string); ok && v != "" {
			target = v // records | trajectory | both
		}

		// 解析时间
		start, err := time.Parse(time.RFC3339, req.StartTime)
		if err != nil {
			http.Error(w, "开始时间格式错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		end, err := time.Parse(time.RFC3339, req.EndTime)
		if err != nil {
			http.Error(w, "结束时间格式错误: "+err.Error(), http.StatusBadRequest)
			return
		}
		start = start.UTC()
		end = end.UTC()

		duration := end.Sub(start)
		useInflux := duration <= 72*time.Hour

		// 临时文件名
		tmpDir := os.TempDir()
		recordFile := filepath.Join(tmpDir, "flightRecord.xlsx")
		trajFile := filepath.Join(tmpDir, "flightTrajectory.xlsx")
		var toServePath string
		var serveName string

		// helper: export records
		exportRecords := func() error {
			// 尝试使用 influx（仅当 useInflux）否则使用 MySQL
			if useInflux && svcCtx.InfluxDao != nil {
				recs, err := svcCtx.InfluxDao.GetFlightDate(start, end)
				if err == nil && len(recs) > 0 {
					return dao.ExportFlightRecordsToExcel(recs, recordFile)
				}
				// 若 influx 失败或无数据，回退到 MySQL
			}
			// 使用 MySQL
			if svcCtx.MySQLDao == nil {
				return fmtError("MySQL 未配置，无法导出")
			}
			// MySQL 查询需要格式化时间
			st := start.Format("2006-01-02 15:04:05")
			ed := end.Format("2006-01-02 15:04:05")
			recs, err := svcCtx.MySQLDao.QueryFlightRecords(req.OrderID, req.UasID, st, ed)
			if err != nil {
				return err
			}
			return dao.ExportFlightRecordsToExcel(recs, recordFile)
		}

		// helper: export trajectory (always from MySQL)
		exportTrajectory := func() error {
			if svcCtx.MySQLDao == nil {
				return fmtError("MySQL 未配置，无法导出轨迹")
			}
			st := start.Format("2006-01-02 15:04:05")
			ed := end.Format("2006-01-02 15:04:05")
			pts, err := svcCtx.MySQLDao.QueryTrackPoints(st, ed, req.OrderID)
			if err != nil {
				return err
			}
			return dao.ExportFlightRecordsToExcel(pts, trajFile)
		}

		// 执行导出
		switch target {
		case "records":
			if err := exportRecords(); err != nil {
				http.Error(w, "导出 records 失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			toServePath = recordFile
			serveName = "flightRecord.xlsx"
		case "trajectory":
			if err := exportTrajectory(); err != nil {
				http.Error(w, "导出 trajectory 失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			toServePath = trajFile
			serveName = "flightTrajectory.xlsx"
		case "both":
			if err := exportRecords(); err != nil {
				http.Error(w, "导出 records 失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			if err := exportTrajectory(); err != nil {
				http.Error(w, "导出 trajectory 失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			zipFile := filepath.Join(tmpDir, "flight_export.zip")
			if err := export.CreateZip([]string{recordFile, trajFile}, zipFile); err != nil {
				http.Error(w, "打包失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			toServePath = zipFile
			serveName = "flight_export.zip"
			// cleanup zip later
			defer os.Remove(zipFile)
		default:
			http.Error(w, "unknown target parameter", http.StatusBadRequest)
			return
		}

		// 清理导出临时文件
		defer func() {
			_ = os.Remove(recordFile)
			_ = os.Remove(trajFile)
		}()

		// 设置响应头并返回文件
		if serveName == "flight_export.zip" {
			w.Header().Set("Content-Type", "application/zip")
		} else {
			w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+serveName)
		http.ServeFile(w, r, toServePath)
	}
}

// 简单的错误构造器，避免引入 fmt 在轻量场景
func fmtError(msg string) error {
	return &simpleError{msg}
}

type simpleError struct{ s string }

func (e *simpleError) Error() string { return e.s }
