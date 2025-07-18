package model

import "time"

type FlightRecord struct {
	ID          int       `db:"id"`
	OrderID     string    `db:"orderID"`
	UasID       string    `db:"uasID"` // 对应无人机编号，UAS04028624 == 5197, UAS04143500 == 5210, UAS04028648 == 5203
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	StartLat    int64     `db:"start_lat"`
	StartLng    int64     `db:"start_lng"`
	EndLat      int64     `db:"end_lat"`
	EndLng      int64     `db:"end_lng"`
	Distance    float64   `db:"distance"`
	BatteryUsed float64   `db:"battery_used"`
	CreatedAt   time.Time `db:"created_at"`
	Payload     float64   `db:"payload"`
}

type FlightTrackPoint struct {
	ID           int64     `db:"id"`
	OrderID      string    `db:"orderID"`
	FlightStatus string    `db:"flightStatus"`
	TimeStamp    time.Time `db:"timeStamp"`
	Longitude    int64     `db:"longitude"`
	Latitude     int64     `db:"latitude"`
	HeightType   int       `db:"heightType"`
	Height       int       `db:"height"`
	Altitude     int       `db:"altitude"`
	VS           int       `db:"VS"`
	GS           int       `db:"GS"`
	Course       int       `db:"course"`
	SOC          int       `db:"SOC"`
	RM           int       `db:"RM"`
	Voltage      int       `db:"voltage"`
	Current      int       `db:"current"`
	WindSpeed    int       `db:"windSpeed"`
	WindDirect   int       `db:"windDirect"`
	Temperture   int       `db:"temperture"`
	Humidity     int       `db:"humidity"`
}
