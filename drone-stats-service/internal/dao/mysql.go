package dao

import (
	"database/sql"
	"drone-stats-service/internal/model"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDao struct {
	DB *sql.DB
}

func NewMySQLDao(dsn string) (*MySQLDao, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20) // 适当调大
	db.SetMaxIdleConns(10)
	return &MySQLDao{DB: db}, nil
}

// 保存飞行记录
func (d *MySQLDao) SaveFlightRecord(uavId string, startTime, endTime time.Time, start_lat, start_lng, end_lat, end_lng int64, distance float64, batteryUsed int) error {
	_, err := d.DB.Exec(
		`INSERT INTO flight_records (
			uav_id,
			start_time,
			end_time,
			start_lat,
			start_lng,
			end_lat,
			end_lng,
			distance,
			battery_used,
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		uavId, startTime, endTime, start_lat, start_lng, end_lat, end_lng, distance, batteryUsed)
	return err
}

// 保存主表并返回自增ID
func (d *MySQLDao) SaveFlightRecordAndGetID(fr model.FlightRecord) (int64, error) {
	res, err := d.DB.Exec(`INSERT INTO flight_records 
        (uav_id, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used) 
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fr.UavId, fr.StartTime, fr.EndTime, fr.StartLat, fr.StartLng, fr.EndLat, fr.EndLng, fr.Distance, fr.BatteryUsed)
	if err != nil {
		fmt.Println("MySQL主表写入错误:", err)
		return 0, err
	}
	// fmt.Println("飞行记录写入MySQL成功")
	return res.LastInsertId()
}

// 保存轨迹点
func (d *MySQLDao) SaveTrackPoints(points []model.FlightTrackPoint) error {
	if len(points) == 0 {
		return nil
	}
	fmt.Println("MySQL子表开始写入")
	query := "INSERT INTO flight_track_points (flight_record_id, flight_status, time_stamp, longitude, latitude, altitude, soc) VALUES "
	vals := []interface{}{}
	for _, tp := range points {
		query += "(?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals, tp.FlightRecordID, tp.FlightStatus, tp.TimeStamp.Format("2006-01-02 15:04:05"),
			tp.Longitude, tp.Latitude, tp.Altitude, tp.SOC)
	}
	query = query[:len(query)-1] // 去掉最后一个逗号
	_, err := d.DB.Exec(query, vals...)
	if err != nil {
		fmt.Println("批量插入轨迹点失败:", err)
	} else {
		fmt.Println("飞行轨迹写入MySQL成功")
	}
	return err
}

// 查询总无人机数
func (d *MySQLDao) CountTotalUas() (int, error) {
	var total int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM uas_devices").Scan(&total)
	return total, err
}

// 查询在线无人机数（假设status=1为在线）
func (d *MySQLDao) CountOnlineUas() (int, error) {
	var online int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM uas_devices WHERE status=1").Scan(&online)
	return online, err
}

// 注册新无人机
func (d *MySQLDao) RegisterUasIfNotExist(uasId string, regTime time.Time) error {
	var exists int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM uas_devices WHERE uas_id=?", uasId).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		_, err := d.DB.Exec("INSERT INTO uas_devices (uas_id, register_time, status) VALUES (?, ?, ?)", uasId, regTime, 0)
		return err
	}
	return nil
}

// 判断飞行架次是否已存在
func (d *MySQLDao) FlightRecordExists(uavId string, startTime, endTime time.Time) (bool, error) {
	var cnt int
	err := d.DB.QueryRow(
		"SELECT COUNT(*) FROM flight_records WHERE uav_id=? AND start_time=? AND end_time=?",
		uavId, startTime, endTime,
	).Scan(&cnt)
	return cnt > 0, err
}
