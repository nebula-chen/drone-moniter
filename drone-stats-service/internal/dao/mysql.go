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
func (d *MySQLDao) SaveFlightRecord(orderID string, startTime, endTime time.Time, start_lat, start_lng, end_lat, end_lng int64, distance float64, batteryUsed int) error {
	_, err := d.DB.Exec(
		`INSERT INTO flight_records (
			orderID,
			start_time,
			end_time,
			start_lat,
			start_lng,
			end_lat,
			end_lng,
			distance,
			battery_used,
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, startTime, endTime, start_lat, start_lng, end_lat, end_lng, distance, batteryUsed)
	return err
}

// 保存主表并返回orderID（飞行架次唯一编号）
func (d *MySQLDao) SaveFlightRecordAndGetOrderID(fr model.FlightRecord) (string, error) {
	_, err := d.DB.Exec(`INSERT INTO flight_records 
		(orderID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, payload) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fr.OrderID, fr.StartTime, fr.EndTime, fr.StartLat, fr.StartLng, fr.EndLat, fr.EndLng, fr.Distance, fr.BatteryUsed, fr.Payload)
	if err != nil {
		fmt.Println("MySQL主表写入错误:", err)
		return "", err
	}
	return fr.OrderID, nil
}

// 保存轨迹点
func (d *MySQLDao) SaveTrackPoints(points []model.FlightTrackPoint) error {
	if len(points) == 0 {
		return nil
	}
	query := "INSERT INTO flight_track_points (orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, windSpeed, windDirect, temperture, humidity) VALUES "
	vals := []interface{}{}
	for _, tp := range points {
		query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals,
			tp.OrderID, tp.FlightStatus, tp.TimeStamp.Format("2006-01-02 15:04:05"),
			tp.Longitude, tp.Latitude, tp.HeightType, tp.Height, tp.Altitude, tp.VS, tp.GS, tp.Course, tp.SOC, tp.RM, tp.WindSpeed, tp.WindDirect, tp.Temperture, tp.Humidity)
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
func (d *MySQLDao) CountTotalSorties() (int, error) {
	var total int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM flight_sorties").Scan(&total)
	return total, err
}

// 查询在线无人机数（假设status=1为在线）
func (d *MySQLDao) CountOnlineSorties() (int, error) {
	var online int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM flight_sorties WHERE status=1").Scan(&online)
	return online, err
}

// 注册新架次
func (d *MySQLDao) RegisterSortiesIfNotExist(orderID string, regTime time.Time) error {
	var exists int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM flight_sorties WHERE OrderID=?", orderID).Scan(&exists)
	if err != nil {
		return err
	}
	if exists == 0 {
		_, err := d.DB.Exec("INSERT INTO flight_sorties (OrderID, register_time) VALUES (?, ?)", orderID, regTime)
		return err
	}
	return nil
}

// 判断飞行架次是否已存在
func (d *MySQLDao) FlightRecordExists(orderID string, startTime, endTime time.Time) (bool, error) {
	var cnt int
	err := d.DB.QueryRow(
		"SELECT COUNT(*) FROM flight_records WHERE orderID=? AND start_time=? AND end_time=?",
		orderID, startTime, endTime,
	).Scan(&cnt)
	return cnt > 0, err
}

// 查询飞行记录（支持条件筛选）
func (d *MySQLDao) QueryFlightRecords(orderID, startTime, endTime string) ([]map[string]interface{}, error) {
	query := `SELECT id, orderID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, created_at
        FROM flight_records WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND orderID=?"
		args = append(args, orderID)
	}
	if startTime != "" {
		query += " AND start_time >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND end_time <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY start_time DESC LIMIT 100"
	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var records []map[string]interface{}
	for rows.Next() {
		var (
			id, batteryUsed                    int
			orderID                            string
			startTime, endTime, createdAt      sql.NullTime
			startLat, startLng, endLat, endLng sql.NullInt64
			distance                           sql.NullFloat64
		)
		err := rows.Scan(&id, &orderID, &startTime, &endTime, &startLat, &startLng, &endLat, &endLng, &distance, &batteryUsed, &createdAt)
		if err != nil {
			continue
		}
		record := map[string]interface{}{
			"id":           id,
			"OrderID":      orderID,
			"start_time":   startTime.Time.Format("2006-01-02 15:04:05"),
			"end_time":     endTime.Time.Format("2006-01-02 15:04:05"),
			"start_lat":    startLat.Int64,
			"start_lng":    startLng.Int64,
			"end_lat":      endLat.Int64,
			"end_lng":      endLng.Int64,
			"distance":     distance.Float64,
			"battery_used": batteryUsed,
			"created_at":   createdAt.Time.Format("2006-01-02 15:04:05"),
		}
		records = append(records, record)
	}
	return records, nil
}

// 统计总飞行架次、总航程、总飞行时长（单位：秒）
func (d *MySQLDao) GetFlightStats() (totalCount int, totalDistance float64, totalTime int64, err error) {
	rows, err := d.DB.Query(`
        SELECT start_time, end_time, distance FROM flight_records
    `)
	if err != nil {
		return
	}
	defer rows.Close()
	var (
		startTime, endTime sql.NullTime
		distance           sql.NullFloat64
	)
	for rows.Next() {
		if err = rows.Scan(&startTime, &endTime, &distance); err != nil {
			continue
		}
		totalCount++
		if distance.Valid {
			totalDistance += distance.Float64
		}
		if startTime.Valid && endTime.Valid {
			dur := endTime.Time.Sub(startTime.Time).Seconds()
			if dur > 0 {
				totalTime += int64(dur)
			}
		}
	}
	return
}

// 按年、月、日统计飞行架次
func (d *MySQLDao) GetFlightRecordsStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y') as date, COUNT(*) as count FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var count int
		if err := rows.Scan(&date, &count); err == nil {
			yearStats = append(yearStats, map[string]interface{}{"date": date, "count": count})
		}
	}

	// 月统计
	rows2, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m') as date, COUNT(*) as count FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var date string
		var count int
		if err := rows2.Scan(&date, &count); err == nil {
			monthStats = append(monthStats, map[string]interface{}{"date": date, "count": count})
		}
	}

	// 日统计
	rows3, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m-%d') as date, COUNT(*) as count FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var date string
		var count int
		if err := rows3.Scan(&date, &count); err == nil {
			dayStats = append(dayStats, map[string]interface{}{"date": date, "count": count})
		}
	}
	return
}

// 按年、月、日统计单位距离单位载重耗电量（distance或payload为0时按1处理）
func (d *MySQLDao) GetSOCUsageStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`
        SELECT DATE_FORMAT(start_time, '%Y') as date, 
        SUM(battery_used / 
            (CASE WHEN distance=0 OR distance IS NULL THEN 1 ELSE distance/1000 END) / 
            (CASE WHEN payload=0 OR payload IS NULL THEN 1 ELSE payload END)
        ) as total 
        FROM flight_records 
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var total float64
		if err := rows.Scan(&date, &total); err == nil {
			yearStats = append(yearStats, map[string]interface{}{"date": date, "total": total})
		}
	}

	// 月统计
	rows2, err := d.DB.Query(`
        SELECT DATE_FORMAT(start_time, '%Y-%m') as date, 
        SUM(battery_used / 
            (CASE WHEN distance=0 OR distance IS NULL THEN 1 ELSE distance/1000 END) / 
            (CASE WHEN payload=0 OR payload IS NULL THEN 1 ELSE payload END)
        ) as total 
        FROM flight_records 
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var date string
		var total float64
		if err := rows2.Scan(&date, &total); err == nil {
			monthStats = append(monthStats, map[string]interface{}{"date": date, "total": total})
		}
	}

	// 日统计
	rows3, err := d.DB.Query(`
        SELECT DATE_FORMAT(start_time, '%Y-%m-%d') as date, 
        SUM(battery_used / 
            (CASE WHEN distance=0 OR distance IS NULL THEN 1 ELSE distance/1000 END) / 
            (CASE WHEN payload=0 OR payload IS NULL THEN 1 ELSE payload END)
        ) as total 
        FROM flight_records 
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var date string
		var total float64
		if err := rows3.Scan(&date, &total); err == nil {
			dayStats = append(dayStats, map[string]interface{}{"date": date, "total": total})
		}
	}
	return
}

// 按年、月、日统计运输货量
func (d *MySQLDao) GetPayloadStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y') as date, SUM(payload) as payload FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var payload float64
		if err := rows.Scan(&date, &payload); err == nil {
			yearStats = append(yearStats, map[string]interface{}{"date": date, "payload": payload})
		}
	}

	// 月统计
	rows2, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m') as date, SUM(payload) as payload FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var date string
		var payload float64
		if err := rows2.Scan(&date, &payload); err == nil {
			monthStats = append(monthStats, map[string]interface{}{"date": date, "payload": payload})
		}
	}

	// 日统计
	rows3, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m-%d') as date, SUM(payload) as payload FROM flight_records GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var date string
		var payload float64
		if err := rows3.Scan(&date, &payload); err == nil {
			dayStats = append(dayStats, map[string]interface{}{"date": date, "payload": payload})
		}
	}
	return
}

// 统计平均飞行时长（秒）、平均耗电量、平均速度
func (d *MySQLDao) GetAvgStats() (avgTime float64, avgSOC float64, avgGS float64, err error) {
	var avgTimeNull, avgSOCNull, avgGSNull sql.NullFloat64
	row := d.DB.QueryRow(`
        SELECT 
            AVG(TIMESTAMPDIFF(SECOND, start_time, end_time)) as avg_time,
            AVG(battery_used) as avg_battery,
            (SELECT AVG(gs) FROM flight_track_points WHERE gs IS NOT NULL) as avg_gs
        FROM flight_records
        WHERE end_time IS NOT NULL AND battery_used IS NOT NULL
    `)
	err = row.Scan(&avgTimeNull, &avgSOCNull, &avgGSNull)
	if avgTimeNull.Valid {
		avgTime = avgTimeNull.Float64
	}
	if avgSOCNull.Valid {
		avgSOC = avgSOCNull.Float64
	}
	if avgGSNull.Valid {
		avgGS = avgGSNull.Float64
	}
	return
}

// 查询某条飞行记录的所有轨迹点
func (d *MySQLDao) GetTrackPointsByRecordId(orderID string) ([]map[string]interface{}, error) {
	rows, err := d.DB.Query(`
        SELECT id, orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, windSpeed, windDirect, temperture, humidity
        FROM flight_track_points
        WHERE orderID = ?
        ORDER BY timeStamp ASC
    `, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []map[string]interface{}
	for rows.Next() {
		var (
			id                                int64
			orderID, flightStatus             string
			timeStamp                         time.Time
			longitude, latitude               int64
			heightType, height, altitude      int
			VS, GS, course, SOC, RM           int
			windSpeed, windDirect, temperture int
			humidity                          int
		)
		err := rows.Scan(&id, &orderID, &flightStatus, &timeStamp, &longitude, &latitude, &heightType, &height, &altitude, &VS, &GS, &course, &SOC, &RM, &windSpeed, &windDirect, &temperture, &humidity)
		if err == nil {
			points = append(points, map[string]interface{}{
				"orderID":      orderID,
				"flightStatus": flightStatus,
				"timeStamp":    timeStamp.Format("2006-01-02 15:04:05"),
				"longitude":    longitude,
				"latitude":     latitude,
				"heightType":   heightType,
				"height":       height,
				"altitude":     altitude,
				"VS":           VS,
				"GS":           GS,
				"course":       course,
				"SOC":          SOC,
				"RM":           RM,
				"windSpeed":    windSpeed,
				"windDirect":   windDirect,
				"temperture":   temperture,
				"humidity":     humidity,
			})
		}
	}
	return points, nil
}
