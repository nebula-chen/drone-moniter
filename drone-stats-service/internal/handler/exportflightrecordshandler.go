package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
		format := "xlsx"
		if v, ok := raw["target"].(string); ok && v != "" {
			target = v // records | trajectory | both
		}
		if v, ok := raw["format"].(string); ok && v != "" {
			format = strings.ToLower(v)
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

		// 不再优先使用 Influx，records 一律从 MySQL 导出

		// 临时文件名
		tmpDir := os.TempDir()
		recExt := "xlsx"
		trajExt := "xlsx"
		if format == "csv" {
			recExt = "csv"
			trajExt = "csv"
		}
		recordFile := filepath.Join(tmpDir, "flightRecord."+recExt)
		trajFile := filepath.Join(tmpDir, "flightTrajectory."+trajExt)
		var toServePath string
		var serveName string

		// helper: export records（始终使用 MySQL）
		exportRecords := func() error {
			if svcCtx.MySQLDao == nil {
				return fmtError("MySQL 未配置，无法导出")
			}
			// MySQL 查询需要格式化时间
			st := start.Format("2006-01-02 15:04:05")
			ed := end.Format("2006-01-02 15:04:05")
			// 使用 MySQL 的查询/流式接口导出 records
			if format == "csv" {
				return svcCtx.MySQLDao.ExportFlightRecordsToCSVStream(req.OrderID, req.UasID, st, ed, recordFile)
			}
			return svcCtx.MySQLDao.ExportFlightRecordsToExcelStream(req.OrderID, req.UasID, st, ed, recordFile)
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
			if format == "csv" {
				return export.ExportMapsToCSV(pts, trajFile)
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
			serveName = "flightRecord." + recExt
		case "trajectory":
			if err := exportTrajectory(); err != nil {
				http.Error(w, "导出 trajectory 失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			toServePath = trajFile
			serveName = "flightTrajectory." + trajExt
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
		} else if strings.HasSuffix(serveName, ".csv") {
			w.Header().Set("Content-Type", "text/csv")
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
