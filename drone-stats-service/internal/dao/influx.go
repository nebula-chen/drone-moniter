package dao

import (
	"context"
	"fmt"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
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
