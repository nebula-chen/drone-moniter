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
	headers := []string{}
	if len(records) > 0 {
		for k := range records[0] {
			headers = append(headers, k)
		}
		// 写表头
		for i, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(i+1, 1)
			f.SetCellValue(sheet, cell, h)
		}
		// 写数据
		for rowIdx, rec := range records {
			for colIdx, h := range headers {
				cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
				f.SetCellValue(sheet, cell, rec[h])
			}
		}
	}
	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	return nil
}
