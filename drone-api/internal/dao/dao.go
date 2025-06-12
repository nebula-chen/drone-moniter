package dao

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"drone-api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

type InfluxDao struct {
	InfluxWriter influxdb2.Client
	WriteAPI     api.WriteAPI
}

func NewInfluxDao(client influxdb2.Client, org, bucket string) *InfluxDao {

	writeAPI := client.WriteAPI(org, bucket)

	go func() {
		for dberr := range writeAPI.Errors() {
			fmt.Println("Influxdb write error: ", dberr)
		}
	}()

	return &InfluxDao{
		InfluxWriter: client,
		WriteAPI:     writeAPI,
	}
}

func (d *InfluxDao) AddPoint(point *write.Point) error {
	d.WriteAPI.WritePoint(point)
	return nil
}

func (d *InfluxDao) BuildPoint(droneStatus *types.DroneStatusReq) (*write.Point, error) {

	// 注意时区, 这个时区是UTC+8, 查询时需要转换成UTC
	timeStamp, err := time.Parse(time.RFC3339, droneStatus.TimeStamp)

	// fmt.Println("timeStamp: ", timeStamp)
	if err != nil {
		return nil, err
	}

	point := write.NewPoint("drone_status", // measurement name 相当于表名
		map[string]string{ // Tags, 相当于建立索引
			"recordId": droneStatus.RecordId,
			"uavType":  droneStatus.UavType,
			"uavId":    droneStatus.UavId},
		map[string]interface{}{ // Fields, 相当于表的字段
			"flightTime": droneStatus.FlightTime,
			"longitude":  droneStatus.Longitude,
			"latitude":   droneStatus.Latitude,
			"altitude":   droneStatus.Altitude,
			"height":     droneStatus.Height,
			"course":     droneStatus.Course,
		}, timeStamp)

	return point, nil
}

func (d *InfluxDao) Close() {
	d.WriteAPI.Flush()
	d.InfluxWriter.Close()
}

// 写入飞行记录到 InfluxDB
func (d *InfluxDao) AddFlightRecord(record *types.FlightRecordReq) error {
	logx.Infof("AddFlightRecord called: %+v", record)
	// 解析时间
	startTime, err := time.Parse(time.RFC3339, record.StartTime)
	if err != nil {
		return fmt.Errorf("startTime 解析失败: %w", err)
	}
	endTime, err := time.Parse(time.RFC3339, record.EndTime)
	if err != nil {
		return fmt.Errorf("endTime 解析失败: %w", err)
	}

	point := write.NewPoint("flight_record",
		map[string]string{
			"uavId": record.UavId,
		},
		map[string]interface{}{
			"startLat":    record.StartLat,
			"startLng":    record.StartLng,
			"endLat":      record.EndLat,
			"endLng":      record.EndLng,
			"distance":    record.Distance,
			"batteryUsed": record.BatteryUsed,
			"startTime":   startTime.Unix(),
			"endTime":     endTime.Unix(),
		},
		endTime, // 以结束时间为时间戳
	)
	d.WriteAPI.WritePoint(point)
	logx.Infof("Writing to measurement: flight_record")
	return nil
}

// 查询飞行统计数据
func (d *InfluxDao) HandleFlightStats() (totalFlights int, totalDistance float64, err error) {
	// 构造 Flux 查询
	query := `
        from(bucket: "drone_data")
            |> range(start: 0)
            |> filter(fn: (r) => r._measurement == "flight_record" and r._field == "distance")
            |> group()
            |> sum()
    `
	queryAPI := d.InfluxWriter.QueryAPI("sysu")
	result, err := queryAPI.Query(context.Background(), query)
	if err != nil {
		return 0, 0, err
	}
	defer result.Close()

	totalDistance = 0
	totalFlights = 0
	for result.Next() {
		if result.TableChanged() {
			// 每个表代表一组
		}
		if v, ok := result.Record().Value().(float64); ok {
			totalDistance += v
		}
		totalFlights++
	}
	if result.Err() != nil {
		return 0, 0, result.Err()
	}
	return totalFlights, totalDistance, nil
}

func (d *InfluxDao) QueryFlightRecords(page, pageSize int, startUnix, endUnix int64, uavId string) (records []types.RecordItem, total, totalUav, totalFlights int, err error) {
	fmt.Println("QueryFlightRecords called") // 新增
	// 构造Flux查询
	where := `|> filter(fn: (r) => r._measurement == "flight_record")`
	if startUnix > 0 {
		where += " |> filter(fn: (r) => r.endTime >= " + strconv.FormatInt(startUnix, 10) + ")"
	}
	if endUnix > 0 {
		where += " |> filter(fn: (r) => r.endTime <= " + strconv.FormatInt(endUnix, 10) + ")"
	}
	if uavId != "" {
		where += ` |> filter(fn: (r) => r.uavId == "` + uavId + `")`
	}
	offset := (page - 1) * pageSize
	flux := `
	from(bucket: "drone_data")
	|> range(start: 0)
	` + where + `
	|> sort(columns: ["endTime"], desc: true)
	|> limit(n:` + strconv.Itoa(pageSize) + `, offset:` + strconv.Itoa(offset) + `)
`
	queryAPI := d.InfluxWriter.QueryAPI("sysu")
	result, err := queryAPI.Query(context.Background(), flux)
	if err != nil {
		return
	}
	defer result.Close()

	recordMap := map[string]*types.RecordItem{}
	uavSet := map[string]struct{}{}
	for result.Next() {
		vals := result.Record().Values()
		uav, _ := vals["uavId"].(string)
		if uav != "" {
			uavSet[uav] = struct{}{}
		}
		id := fmt.Sprintf("%v", vals["_time"])
		rec, ok := recordMap[id]
		if !ok {
			rec = &types.RecordItem{
				UavId: uav,
			}
			recordMap[id] = rec
		}
		switch vals["_field"] {
		case "startTime":
			if v, ok := vals["_value"].(string); ok {
				// 解析 RFC3339 字符串为时间戳
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					rec.StartTime = t.Unix()
				}
			}
		case "endTime":
			if v, ok := vals["_value"].(string); ok {
				if t, err := time.Parse(time.RFC3339, v); err == nil {
					rec.EndTime = t.Unix()
				}
			}
		case "distance":
			if v, ok := vals["_value"].(float64); ok {
				rec.Distance = v
			}
		case "batteryUsed":
			if v, ok := vals["_value"].(int64); ok {
				rec.BatteryUsed = int(v)
			} else if v, ok := vals["_value"].(float64); ok {
				rec.BatteryUsed = int(v)
			}
		}
	}
	for _, v := range recordMap {
		records = append(records, *v)
	}
	total = len(records)
	totalUav = len(uavSet)
	totalFlights = total
	return
}
