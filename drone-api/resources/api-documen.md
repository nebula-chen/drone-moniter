无人机websocket协议
* 接口: ws://localhost:19999/api/ws
```
type DroneStatusReq struct {
	RecordId   string  `json:"recordId"`
	UavType    string  `json:"uavType"`
	UavId      string  `json:"uavId"`
	TimeStamp  string  `json:"timeStamp"`  //"YYYY-MM-DD hh:mm:ss" 的格式
	FlightTime int     `json:"flightTime"` // 飞行经过的时间(单位: 秒)
	Longitude  int     `json:"longitude"`  // 实际经纬度*1e7传输
	Latitude   int     `json:"latitude"`
	Altitude   float64 `json:"altitude"`
	Height     float64 `json:"height"`
	Course     float64 `json:"course"`
}
```