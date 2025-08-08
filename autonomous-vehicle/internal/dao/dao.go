package dao

import (
	"fmt"
	"time"

	"autonomous-vehicle/internal/types"

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
	// fmt.Printf("[InfluxDao] 写入数据点: %v\n", point) // 日志输出
	d.WriteAPI.WritePoint(point)
	return nil
}

func (d *InfluxDao) BuildPoint(vehicleInfo *types.VehicleInfo) (*write.Point, error) {
	// fmt.Printf("[InfluxDao] 构建数据点, VIN: %s, 时间戳: %d\n", vehicleInfo.Vin, vehicleInfo.OccurTimestamp) // 日志输出
	// 协议时间戳为毫秒级，需转为 time.Time
	utcTime := time.Unix(vehicleInfo.OccurTimestamp/1000, (vehicleInfo.OccurTimestamp%1000)*int64(time.Millisecond)).UTC()

	point := write.NewPoint("vehicle_info",
		map[string]string{
			"vin":      vehicleInfo.Vin,      // 车辆唯一标识
			"vinId":    vehicleInfo.VinId,    // 车辆ID
			"parkCode": vehicleInfo.ParkCode, // 网格编码
			"parkName": vehicleInfo.ParkName, // 网格名称
		},
		map[string]interface{}{
			"driveMode":      vehicleInfo.DriveMode,      // 驾驶模式（协议：1自动驾驶 2远程 3场景遥控 0缺省）
			"gear":           vehicleInfo.Gear,           // 档位
			"speed":          vehicleInfo.Speed,          // 车速 km/h
			"accelerationV":  vehicleInfo.AccelerationV,  // 纵向加速度，单位0.01m/s2
			"accelerationH":  vehicleInfo.AccelerationH,  // 横向加速度，单位0.01m/s2
			"gnssHead":       vehicleInfo.GnssHead,       // 航向角
			"lon":            vehicleInfo.Position.Lon,   // 经度
			"lat":            vehicleInfo.Position.Lat,   // 纬度
			"electricity":    vehicleInfo.Electricity,    // 后轮电池电量
			"frontBattery":   vehicleInfo.FrontBattery,   // 前轮电池电量
			"realBattery":    vehicleInfo.RealBattery,    // 当前使用电池电量
			"mile":           vehicleInfo.Mile,           // 累计里程 km
			"occurTimestamp": vehicleInfo.OccurTimestamp, // 秒级时间戳
			"powerState":     vehicleInfo.PowerState,     // 是否在线
		}, utcTime)

	return point, nil
}

func (d *InfluxDao) Close() {
	// fmt.Println("[InfluxDao] 关闭 InfluxDB 连接，刷新数据...") // 日志输出
	d.WriteAPI.Flush()
	d.InfluxWriter.Close()
}

func (d *InfluxDao) SaveVehicleInfo(vehicleInfo *types.VehicleInfo) error {
	// fmt.Printf("[InfluxDao] 保存车辆信息, VIN: %s\n", vehicleInfo.Vin) // 日志输出
	point, err := d.BuildPoint(vehicleInfo)
	if err != nil {
		// fmt.Printf("[InfluxDao] 构建数据点失败: %v\n", err)
		return err
	}
	return d.AddPoint(point)
}
