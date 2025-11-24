package export

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExportMapsToExcel 将通用的 map[string]interface{} 列表导出为 xlsx 文件。
// maps 中的键会被当作表头，按 maps[0] 的键顺序写入。
func ExportMapsToExcel(records []map[string]interface{}, filePath string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 如果没有数据，写入空表头（避免客户端报错）
	if len(records) == 0 {
		if err := f.SaveAs(filePath); err != nil {
			return err
		}
		return nil
	}

	// 我们为已知字段提供友好表头并固定列顺序；对于未知字段，附加在末尾
	preferredOrder := []string{"id", "OrderID", "uasID", "start_time", "end_time", "start_lat", "start_lng", "end_lat", "end_lng", "distance", "battery_used", "created_at", "payload", "expressCount"}
	headerNames := map[string]string{
		"id":           "ID",
		"OrderID":      "Order ID",
		"uasID":        "UAS ID",
		"start_time":   "Start Time",
		"end_time":     "End Time",
		"start_lat":    "Start Latitude",
		"start_lng":    "Start Longitude",
		"end_lat":      "End Latitude",
		"end_lng":      "End Longitude",
		"distance":     "Distance (m)",
		"battery_used": "Battery Used (A.h)",
		"created_at":   "Created At",
		"payload":      "Payload (kg)",
		"expressCount": "Express Count",
	}

	// 构造列顺序
	headers := make([]string, 0, len(records[0]))
	seen := map[string]bool{}
	for _, k := range preferredOrder {
		if _, ok := records[0][k]; ok {
			headers = append(headers, k)
			seen[k] = true
		}
	}
	// 追加其他字段
	for k := range records[0] {
		if !seen[k] {
			headers = append(headers, k)
		}
	}

	// 写表头（友好名称）
	for i, key := range headers {
		title := key
		if n, ok := headerNames[key]; ok {
			title = n
		}
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, title)
	}

	// 写数据并格式化已知字段
	for rIdx, rec := range records {
		for cIdx, key := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			val := formatFieldForExcel(key, rec[key])
			f.SetCellValue(sheet, cell, val)
		}
	}
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}

// ExportMapsToCSV 将通用的 map[string]interface{} 列表导出为 CSV 文件。
// maps 中的键会被当作表头，按 records[0] 的键顺序写入。
func ExportMapsToCSV(records []map[string]interface{}, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	defer bw.Flush()
	w := csv.NewWriter(bw)
	defer w.Flush()

	if len(records) == 0 {
		// 写空文件
		return nil
	}

	// 构造表头顺序
	headers := make([]string, 0, len(records[0]))
	for k := range records[0] {
		headers = append(headers, k)
	}
	if err := w.Write(headers); err != nil {
		return err
	}

	// 写数据
	for _, rec := range records {
		row := make([]string, 0, len(headers))
		for _, h := range headers {
			v := rec[h]
			var s string
			if v == nil {
				s = ""
			} else {
				switch t := v.(type) {
				case string:
					s = t
				case time.Time:
					s = t.Format("2006-01-02 15:04:05")
				case float64:
					s = strconv.FormatFloat(t, 'f', -1, 64)
				case float32:
					s = strconv.FormatFloat(float64(t), 'f', -1, 32)
				case int:
					s = strconv.Itoa(t)
				case int64:
					s = strconv.FormatInt(t, 10)
				case uint64:
					s = strconv.FormatUint(t, 10)
				case bool:
					if t {
						s = "true"
					} else {
						s = "false"
					}
				default:
					s = fmt.Sprint(v)
				}
			}
			row = append(row, s)
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// formatFieldForExcel 对常见字段做更友好的格式化
func formatFieldForExcel(key string, v interface{}) interface{} {
	if v == nil {
		return ""
	}
	switch key {
	case "start_time", "end_time", "created_at", "timeStamp":
		switch t := v.(type) {
		case string:
			return t
		case time.Time:
			return t.Format("2006-01-02 15:04:05")
		default:
			return v
		}
	case "start_lat", "start_lng", "end_lat", "end_lng", "longitude", "latitude":
		switch n := v.(type) {
		case int64:
			return float64(n) / 1e7
		case int:
			return float64(n) / 1e7
		case float64:
			return n / 1e7
		case string:
			return n
		default:
			return v
		}
	case "VS", "GS":
		switch n := v.(type) {
		case int:
			return float64(n) / 10.0
		case int64:
			return float64(n) / 10.0
		case float64:
			return n / 10.0
		default:
			return v
		}
	case "payload":
		switch n := v.(type) {
		case int:
			return float64(n) / 10.0
		case int64:
			return float64(n) / 10.0
		case float64:
			return n / 10.0
		default:
			return v
		}
	default:
		return v
	}
}

// CreateZip 包装多个文件为一个 zip 并返回 zipPath
func CreateZip(files []string, zipPath string) error {
	zf, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zf.Close()
	zw := zip.NewWriter(zf)
	defer zw.Close()
	for _, p := range files {
		fname := filepath.Base(p)
		w, err := zw.Create(fname)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
	return nil
}
