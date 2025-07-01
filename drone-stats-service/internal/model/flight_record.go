package model

import "time"

type FlightRecord struct {
	ID          int       `db:"id"`
	OrderID     string    `db:"orderID"`
	StartTime   time.Time `db:"start_time"`
	EndTime     time.Time `db:"end_time"`
	StartLat    int64     `db:"start_lat"`
	StartLng    int64     `db:"start_lng"`
	EndLat      int64     `db:"end_lat"`
	EndLng      int64     `db:"end_lng"`
	Distance    float64   `db:"distance"`
	BatteryUsed int       `db:"battery_used"`
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
	WindSpeed    int       `db:"windSpeed"`
	WindDirect   int       `db:"windDirect"`
	Temperture   int       `db:"temperture"`
	Humidity     int       `db:"humidity"`
}
