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

	// 注意时区, 这个时区是UTC+8, 需要转换成UTC0
	timeStamp, err := time.Parse(time.RFC3339, droneStatus.TimeStamp)

	// fmt.Println("timeStamp: ", timeStamp)
	if err != nil {
		return nil, err
	}
	utcTime := timeStamp.Add(-8 * time.Hour)

	point := write.NewPoint("drone_status", // measurement name 相当于表名
		map[string]string{ // Tags, 相当于建立索引
			"flightCode": droneStatus.FlightCode,
			"sn":         droneStatus.Sn},
		map[string]interface{}{ // Fields, 相当于表的字段
			"orderID":        droneStatus.OrderID,
			"flightStatus":   droneStatus.FlightStatus,
			"manufacturerID": droneStatus.ManufacturerID,
			"uasID":          droneStatus.UasID,
			"uasModel":       droneStatus.UasModel,
			"coordinate":     droneStatus.Coordinate,
			"longitude":      droneStatus.Longitude,
			"latitude":       droneStatus.Latitude,
			"heightType":     droneStatus.HeightType,
			"height":         droneStatus.Height,
			"altitude":       droneStatus.Altitude,
			"VS":             droneStatus.VS,
			"GS":             droneStatus.GS,
			"course":         droneStatus.Course,
			"SOC":            droneStatus.SOC,
			"RM":             droneStatus.RM,
			"windSpeed":      droneStatus.WindSpeed,
			"windDirect":     droneStatus.WindDirect,
			"temperture":     droneStatus.Temperture,
			"humidity":       droneStatus.Humidity,
		}, utcTime)

	return point, nil
}

func (d *InfluxDao) Close() {
	d.WriteAPI.Flush()
	d.InfluxWriter.Close()
}
