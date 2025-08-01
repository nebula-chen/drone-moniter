package dao

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	d.WriteAPI.WritePoint(point)
	return nil
}

func (d *InfluxDao) BuildPoint(vehicleInfo *types.VehicleInfo) (*write.Point, error) {
	// 协议时间戳为秒级，需转为 time.Time
	utcTime := time.Unix(vehicleInfo.OccurTimestamp, 0).UTC()

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
	d.WriteAPI.Flush()
	d.InfluxWriter.Close()
}

// 通过 vin 调用 getVehicleInfo 协议接口获取车辆状态
func (d *InfluxDao) GetVehicleInfoByVin(
	vin string,
	xFrom string,
	xVersion string,
	genSignParams func() (string, string, string, string, error),
) (*types.VehicleInfo, error) {
	timestamp, nonce, signature, token, err := genSignParams()
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://scapi.test.neolix.net/openapi-server/slvapi/getVehicleInfo?signature=%s&timeStamp=%s&nonce=%s&access_token=%s&vin=%s",
		signature, timestamp, nonce, token, vin)

	httpReq, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("X-From", xFrom)
	httpReq.Header.Set("X-Version", xVersion)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var result types.GetVehicleInfoResp
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	if result.Code != "10000" {
		return nil, fmt.Errorf("getVehicleInfo failed: %s", result.Msg)
	}
	return &result.Data, nil
}

func (d *InfluxDao) SaveVehicleInfo(vehicleInfo *types.VehicleInfo) error {
	point, err := d.BuildPoint(vehicleInfo)
	if err != nil {
		return err
	}
	return d.AddPoint(point)
}
