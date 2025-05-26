package dao

import (
	"fmt"
	"time"

	"drone-api/internal/types"

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
