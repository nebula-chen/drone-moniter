package dao

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/xuri/excelize/v2"
)

type InfluxDao struct {
	Client   influxdb2.Client
	QueryAPI api.QueryAPI
}

func NewInfluxDao(client influxdb2.Client, org string) *InfluxDao {
	return &InfluxDao{
		Client:   client,
		QueryAPI: client.QueryAPI(org),
	}
}

// 查询指定无人机在时间范围内的飞行数据
func (d *InfluxDao) QueryFlightRecords(orderID string, start, end time.Time) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
		from(bucket:"drone_data")
		|> range(start: %s, stop: %s)
		|> filter(fn: (r) => r["_measurement"] == "drone_status")
		|> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		|> filter(fn: (r) => r["orderID"] == "%s")
		`, start.Format(time.RFC3339), end.Format(time.RFC3339), orderID,
	)

	result, err := d.QueryAPI.Query(context.Background(), query)
	if err != nil {
		fmt.Println("InfluxDB查询报错:", err)
		return nil, err
	}
	// fmt.Println("InfluxDB查询完成:", err)
	var records []map[string]interface{}
	for result.Next() {
		records = append(records, result.Record().Values())
	}
	return records, result.Err()
}

// 查询所有无人机ID及首次出现时间
func (d *InfluxDao) GetAllUasIDsAndFirstSeen() (map[string]time.Time, error) {
	query := `
        from(bucket:"drone_data")
        |> range(start: 0)
        |> filter(fn: (r) => r._measurement == "drone_status")
        |> filter(fn: (r) => r._field == "orderID")
        |> keep(columns: ["_time", "_value"])
        |> sort(columns: ["_time"], desc: false)
    `
	result, err := d.QueryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]time.Time)
	for result.Next() {
		id, ok := result.Record().Value().(string)
		if !ok {
			continue
		}
		t, ok := result.Record().ValueByKey("_time").(time.Time)
		if !ok {
			continue
		}
		// 只保留最早时间
		if _, exists := ids[id]; !exists {
			ids[id] = t
		}
	}
	return ids, nil
}

// 拉取指定时间范围的飞行数据
func (d *InfluxDao) GetFlightDate(start, end time.Time) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`
        from(bucket:"drone_data")
        |> range(start: %s, stop: %s)
        |> filter(fn: (r) => r._measurement == "drone_status")
        |> pivot(rowKey:["_time"], columnKey: ["_field"], valueColumn: "_value")
		`, start.Format(time.RFC3339), end.Format(time.RFC3339),
	)
	result, err := d.QueryAPI.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}
	var records []map[string]interface{}
	for result.Next() {
		records = append(records, result.Record().Values())
	}
	return records, result.Err()
}

// 导出为 Excel
func ExportFlightRecordsToExcel(records []map[string]interface{}, filePath string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"
	if len(records) == 0 {
		if err := f.SaveAs(filePath); err != nil {
			return err
		}
		return nil
	}
	// 采用与通用导出相同的友好表头策略（简化版）：按 records[0] 的字段顺序写入表头
	headers := make([]string, 0, len(records[0]))
	for k := range records[0] {
		headers = append(headers, k)
	}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	// 写数据并做基础格式化（时间/经纬度）
	for rIdx, rec := range records {
		for cIdx, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rIdx+2)
			v := rec[h]
			// 简单格式化常见字段
			switch h {
			case "start_time", "end_time", "created_at", "timeStamp":
				switch t := v.(type) {
				case time.Time:
					f.SetCellValue(sheet, cell, t.Format("2006-01-02 15:04:05"))
					continue
				}
			case "start_lat", "start_lng", "end_lat", "end_lng", "longitude", "latitude":
				switch n := v.(type) {
				case int64:
					f.SetCellValue(sheet, cell, float64(n)/1e7)
					continue
				case float64:
					f.SetCellValue(sheet, cell, n/1e7)
					continue
				}
			}
			f.SetCellValue(sheet, cell, v)
		}
	}
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}
