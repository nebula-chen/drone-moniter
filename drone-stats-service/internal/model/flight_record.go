package model

import "time"

type FlightRecord struct {
	ID          int       `db:"id"`
	UavId       string    `db:"uav_id"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	StartLat    int64     `db:"start_lat"`
	StartLng    int64     `db:"start_lng"`
	EndLat      int64     `db:"end_lat"`
	EndLng      int64     `db:"end_lng"`
	Distance    float64   `db:"distance"`
	BatteryUsed int       `db:"battery_used"`
}

type FlightTrackPoint struct {
	ID             int64
	FlightRecordID int64
	FlightStatus   string
	TimeStamp      time.Time
	Longitude      float64
	Latitude       float64
	Altitude       float64
	SOC            int
	GS             float64 // 新增字段：速度
}
