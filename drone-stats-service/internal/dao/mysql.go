package dao

import (
	"bufio"
	"database/sql"
	"drone-stats-service/internal/model"
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/xuri/excelize/v2"
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
func (d *MySQLDao) SaveFlightRecord(orderID, uasID string, startTime, endTime time.Time, start_lat, start_lng, end_lat, end_lng int64, distance, batteryUsed float64) error {
	_, err := d.DB.Exec(
		`INSERT INTO flight_records (
			orderID,
			uasID,
			start_time,
			end_time,
			start_lat,
			start_lng,
			end_lat,
			end_lng,
			distance,
			battery_used
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, uasID, startTime, endTime, start_lat, start_lng, end_lat, end_lng, distance, batteryUsed)
	return err
}

// 保存主表并返回orderID（飞行架次唯一编号）
func (d *MySQLDao) SaveFlightRecordAndGetOrderID(fr model.FlightRecord) (string, error) {
	_, err := d.DB.Exec(`INSERT INTO flight_records 
		(orderID, uasID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, payload) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		fr.OrderID, fr.UasID, fr.StartTime, fr.EndTime, fr.StartLat, fr.StartLng, fr.EndLat, fr.EndLng, fr.Distance, fr.BatteryUsed, fr.Payload)
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
	query :=
		`INSERT INTO flight_track_points (
			orderID,
			flightStatus,
			timeStamp,
			longitude,
			latitude,
			heightType,
			height,
			altitude,
			VS,
			GS,
			course,
			SOC,
			RM,
			voltage,
			current,
			windSpeed,
			windDirect,
			temperture,
			humidity
		) VALUES `
	vals := []interface{}{}
	for _, tp := range points {
		query += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"
		vals = append(vals,
			tp.OrderID, tp.FlightStatus, tp.TimeStamp.Format("2006-01-02 15:04:05"),
			tp.Longitude, tp.Latitude, tp.HeightType, tp.Height, tp.Altitude, tp.VS, tp.GS, tp.Course, tp.SOC, tp.RM, tp.Voltage, tp.Current, tp.WindSpeed, tp.WindDirect, tp.Temperture, tp.Humidity)
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
func (d *MySQLDao) QueryFlightRecords(orderID, uasID, startTime, endTime string) ([]map[string]interface{}, error) {
	query := `SELECT id, OrderID, uasID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, created_at, payload, expressCount
        FROM flight_records WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND OrderID=?"
		args = append(args, orderID)
	}
	if uasID != "" {
		query += " AND uasID=? AND start_lat < 228000000 AND start_lng > 1139430000"
		args = append(args, uasID)
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
			id, payload, expressCount          int
			orderID, uasID                     string
			startTime, endTime, createdAt      sql.NullTime
			startLat, startLng, endLat, endLng sql.NullInt64
			distance, batteryUsed              sql.NullFloat64
		)
		err := rows.Scan(&id, &orderID, &uasID, &startTime, &endTime, &startLat, &startLng, &endLat, &endLng, &distance, &batteryUsed, &createdAt, &payload, &expressCount)
		if err != nil {
			continue
		}
		record := map[string]interface{}{
			"id":           id,
			"OrderID":      orderID,
			"uasID":        uasID,
			"start_time":   startTime.Time.Format("2006-01-02 15:04:05"),
			"end_time":     endTime.Time.Format("2006-01-02 15:04:05"),
			"start_lat":    startLat.Int64,
			"start_lng":    startLng.Int64,
			"end_lat":      endLat.Int64,
			"end_lng":      endLng.Int64,
			"distance":     distance.Float64,
			"battery_used": batteryUsed.Float64,
			"created_at":   createdAt.Time.Format("2006-01-02 15:04:05"),
			"payload":      payload,
			"expressCount": expressCount,
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

// 按年、月、日统计净电能（battery_used总和）
func (d *MySQLDao) GetSOCUsageStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`
        SELECT DATE_FORMAT(start_time, '%Y') as date, 
        SUM(battery_used) as total 
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
        SUM(battery_used) as total 
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
        SUM(battery_used) as total 
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

// 按年、月、日统计总电能/总距离/总载重（distance或payload为0时正常处理，为null按0处理）
func (d *MySQLDao) GetAvgSOCPerDistancePayloadStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`
        SELECT 
            DATE_FORMAT(start_time, '%Y') as date,
            SUM(battery_used) as total_battery,
            SUM(IFNULL(distance,0)/1000) as total_distance,
            SUM(IFNULL(payload,0)/10) as total_payload
        FROM flight_records
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var date string
		var totalBattery, totalDistance, totalPayload float64
		if err := rows.Scan(&date, &totalBattery, &totalDistance, &totalPayload); err == nil {
			var avg float64
			if totalDistance != 0 && totalPayload != 0 {
				avg = totalBattery / totalDistance / totalPayload
			}
			yearStats = append(yearStats, map[string]interface{}{"date": date, "avg": avg})
		}
	}

	// 月统计
	rows2, err := d.DB.Query(`
        SELECT 
            DATE_FORMAT(start_time, '%Y-%m') as date,
            SUM(battery_used) as total_battery,
            SUM(IFNULL(distance,0)/1000) as total_distance,
            SUM(IFNULL(payload,0)/10) as total_payload
        FROM flight_records
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows2.Close()
	for rows2.Next() {
		var date string
		var totalBattery, totalDistance, totalPayload float64
		if err := rows2.Scan(&date, &totalBattery, &totalDistance, &totalPayload); err == nil {
			var avg float64
			if totalDistance != 0 && totalPayload != 0 {
				avg = totalBattery / totalDistance / totalPayload
			}
			monthStats = append(monthStats, map[string]interface{}{"date": date, "avg": avg})
		}
	}

	// 日统计
	rows3, err := d.DB.Query(`
        SELECT 
            DATE_FORMAT(start_time, '%Y-%m-%d') as date,
            SUM(battery_used) as total_battery,
            SUM(IFNULL(distance,0)/1000) as total_distance,
            SUM(IFNULL(payload,0)/10) as total_payload
        FROM flight_records
        GROUP BY date ORDER BY date`)
	if err != nil {
		return
	}
	defer rows3.Close()
	for rows3.Next() {
		var date string
		var totalBattery, totalDistance, totalPayload float64
		if err := rows3.Scan(&date, &totalBattery, &totalDistance, &totalPayload); err == nil {
			var avg float64
			if totalDistance != 0 && totalPayload != 0 {
				avg = totalBattery / totalDistance / totalPayload
			}
			dayStats = append(dayStats, map[string]interface{}{"date": date, "avg": avg})
		}
	}
	return
}

// 按年、月、日统计运输货量
func (d *MySQLDao) GetPayloadStats() (yearStats, monthStats, dayStats []map[string]interface{}, err error) {
	// 年统计
	rows, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y') as date, SUM(payload/10) as payload FROM flight_records GROUP BY date ORDER BY date`)
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
	rows2, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m') as date, SUM(payload/10) as payload FROM flight_records GROUP BY date ORDER BY date`)
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
	rows3, err := d.DB.Query(`SELECT DATE_FORMAT(start_time, '%Y-%m-%d') as date, SUM(payload/10) as payload FROM flight_records GROUP BY date ORDER BY date`)
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

// 统计平均飞行时长（秒）、平均耗电量、平均载货量、平均速度
func (d *MySQLDao) GetAvgStats() (avgTime float64, avgSOC float64, avgPayload float64, avgGS float64, err error) {
	var avgTimeNull, avgEnergyNull, avgPayloadNull, avgGSNull sql.NullFloat64
	row := d.DB.QueryRow(`
        SELECT 
            AVG(TIMESTAMPDIFF(SECOND, start_time, end_time)) as avg_time,
            AVG(battery_used) as avg_battery,
            AVG(CASE WHEN payload=0 OR payload IS NULL THEN NULL ELSE payload/10 END) as avg_payload,
            (SELECT AVG(gs/10) FROM flight_track_points WHERE gs IS NOT NULL) as avg_gs
        FROM flight_records
        WHERE end_time IS NOT NULL AND battery_used IS NOT NULL
    `)
	err = row.Scan(&avgTimeNull, &avgEnergyNull, &avgPayloadNull, &avgGSNull)
	if avgTimeNull.Valid {
		avgTime = avgTimeNull.Float64
	}
	if avgEnergyNull.Valid {
		avgSOC = avgEnergyNull.Float64
	}
	if avgPayloadNull.Valid {
		avgPayload = avgPayloadNull.Float64
	}
	if avgGSNull.Valid {
		avgGS = avgGSNull.Float64
	}
	return
}

// 查询某条飞行记录的所有轨迹点
func (d *MySQLDao) GetTrackPointsByRecordId(orderID string) ([]map[string]interface{}, error) {
	rows, err := d.DB.Query(`
        SELECT id, orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, voltage, current, windSpeed, windDirect, temperture, humidity
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
			id                                          int64
			orderID, flightStatus                       string
			timeStamp                                   time.Time
			longitude, latitude                         int64
			heightType, height, altitude                int
			VS, GS, course, SOC, RM, voltage, current   int
			windSpeed, windDirect, temperture, humidity int
		)
		err := rows.Scan(&id, &orderID, &flightStatus, &timeStamp, &longitude, &latitude, &heightType, &height, &altitude, &VS, &GS, &course, &SOC, &RM, &voltage, &current, &windSpeed, &windDirect, &temperture, &humidity)
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
				"voltage":      voltage,
				"current":      current,
				"windSpeed":    windSpeed,
				"windDirect":   windDirect,
				"temperture":   temperture,
				"humidity":     humidity,
			})
		}
	}
	return points, nil
}

// ExportFlightRecordsToExcelStream 使用流式写入将 MySQL 中的 flight_records 导出为 xlsx 文件，减少内存占用
func (d *MySQLDao) ExportFlightRecordsToExcelStream(orderID, uasID, startTime, endTime, filePath string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"
	// 使用流式写入器
	w, err := f.NewStreamWriter(sheet)
	if err != nil {
		return err
	}
	// 固定表头顺序，保证列稳定
	headers := []interface{}{"ID", "Order ID", "UAS ID", "Start Time", "End Time", "Start Latitude", "Start Longitude", "End Latitude", "End Longitude", "Distance (m)", "Battery Used (kWh)", "Created At", "Payload (kg)", "Express Count"}
	if err := w.SetRow("A1", headers); err != nil {
		return err
	}

	query := `SELECT id, OrderID, uasID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, created_at, payload, expressCount FROM flight_records WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND OrderID=?"
		args = append(args, orderID)
	}
	if uasID != "" {
		query += " AND uasID=?"
		args = append(args, uasID)
	}
	if startTime != "" {
		query += " AND start_time >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND end_time <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY start_time DESC"

	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	rowIdx := 2
	for rows.Next() {
		var (
			id, payload, expressCount          int
			orderIDs, uasIDs                   string
			startTimeT, endTimeT, createdAt    sql.NullTime
			startLat, startLng, endLat, endLng sql.NullInt64
			distance, batteryUsed              sql.NullFloat64
		)
		if err := rows.Scan(&id, &orderIDs, &uasIDs, &startTimeT, &endTimeT, &startLat, &startLng, &endLat, &endLng, &distance, &batteryUsed, &createdAt, &payload, &expressCount); err != nil {
			continue
		}
		// 对某些字段做格式化：经纬度除以1e7，payload除以10
		var startLatVal interface{}
		if startLat.Valid {
			startLatVal = float64(startLat.Int64) / 1e7
		} else {
			startLatVal = ""
		}
		var startLngVal interface{}
		if startLng.Valid {
			startLngVal = float64(startLng.Int64) / 1e7
		} else {
			startLngVal = ""
		}
		var endLatVal interface{}
		if endLat.Valid {
			endLatVal = float64(endLat.Int64) / 1e7
		} else {
			endLatVal = ""
		}
		var endLngVal interface{}
		if endLng.Valid {
			endLngVal = float64(endLng.Int64) / 1e7
		} else {
			endLngVal = ""
		}
		var payloadVal interface{}
		payloadVal = float64(payload) / 10.0

		vals := []interface{}{
			id,
			orderIDs,
			uasIDs,
			nullableTimeFormat(startTimeT),
			nullableTimeFormat(endTimeT),
			startLatVal,
			startLngVal,
			endLatVal,
			endLngVal,
			nullableFloat64(distance),
			nullableFloat64(batteryUsed),
			nullableTimeFormat(createdAt),
			payloadVal,
			expressCount,
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx)
		if err := w.SetRow(cell, vals); err != nil {
			return err
		}
		rowIdx++
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return f.SaveAs(filePath)
}

// ExportTrackPointsToExcelStream 使用流式写入将 flight_track_points 导出为 xlsx 文件
func (d *MySQLDao) ExportTrackPointsToExcelStream(startTime, endTime, orderID, filePath string) error {
	f := excelize.NewFile()
	sheet := "Sheet1"
	w, err := f.NewStreamWriter(sheet)
	if err != nil {
		return err
	}
	headers := []interface{}{"ID", "Order ID", "Flight Status", "Time Stamp", "Longitude", "Latitude", "Height Type", "Height", "Altitude", "VS (m/s)", "GS (m/s)", "Course", "SOC", "RM", "Voltage", "Current", "Wind Speed", "Wind Direct", "Temperature", "Humidity"}
	if err := w.SetRow("A1", headers); err != nil {
		return err
	}

	query := `SELECT id, orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, voltage, current, windSpeed, windDirect, temperture, humidity FROM flight_track_points WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND orderID = ?"
		args = append(args, orderID)
	}
	if startTime != "" {
		query += " AND timeStamp >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND timeStamp <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY timeStamp ASC"

	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	rowIdx := 2
	for rows.Next() {
		var (
			id                                          int64
			orderIDOut, flightStatus                    string
			timeStamp                                   sql.NullTime
			longitude, latitude                         sql.NullInt64
			heightType, height, altitude                sql.NullInt64
			VS, GS, course, SOC, RM, voltage, current   sql.NullInt64
			windSpeed, windDirect, temperture, humidity sql.NullInt64
		)
		if err := rows.Scan(&id, &orderIDOut, &flightStatus, &timeStamp, &longitude, &latitude, &heightType, &height, &altitude, &VS, &GS, &course, &SOC, &RM, &voltage, &current, &windSpeed, &windDirect, &temperture, &humidity); err != nil {
			continue
		}
		// 格式化经纬度与速度（经纬度/1e7，VS/GS 除以10）
		var lonVal interface{}
		if longitude.Valid {
			lonVal = float64(longitude.Int64) / 1e7
		} else {
			lonVal = ""
		}
		var latVal interface{}
		if latitude.Valid {
			latVal = float64(latitude.Int64) / 1e7
		} else {
			latVal = ""
		}
		var vsVal interface{}
		if VS.Valid {
			vsVal = float64(VS.Int64) / 10.0
		} else {
			vsVal = ""
		}
		var gsVal interface{}
		if GS.Valid {
			gsVal = float64(GS.Int64) / 10.0
		} else {
			gsVal = ""
		}

		vals := []interface{}{
			id,
			orderIDOut,
			flightStatus,
			nullableTimeFormat(timeStamp),
			lonVal,
			latVal,
			nullableInt64(heightType),
			nullableInt64(height),
			nullableInt64(altitude),
			vsVal,
			gsVal,
			nullableInt64(course),
			nullableInt64(SOC),
			nullableInt64(RM),
			nullableInt64(voltage),
			nullableInt64(current),
			nullableInt64(windSpeed),
			nullableInt64(windDirect),
			nullableInt64(temperture),
			nullableInt64(humidity),
		}
		cell, _ := excelize.CoordinatesToCellName(1, rowIdx)
		if err := w.SetRow(cell, vals); err != nil {
			return err
		}
		rowIdx++
	}
	if err := w.Flush(); err != nil {
		return err
	}
	return f.SaveAs(filePath)
}

// ExportFlightRecordsToCSVStream 使用流式写入将 MySQL 中的 flight_records 导出为 csv 文件，减少内存占用
func (d *MySQLDao) ExportFlightRecordsToCSVStream(orderID, uasID, startTime, endTime, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	defer bw.Flush()
	w := csv.NewWriter(bw)
	defer w.Flush()

	// 固定表头顺序，保证列稳定
	headers := []string{"ID", "Order ID", "UAS ID", "Start Time", "End Time", "Start Latitude", "Start Longitude", "End Latitude", "End Longitude", "Distance (m)", "Battery Used (kWh)", "Created At", "Payload (kg)", "Express Count"}
	if err := w.Write(headers); err != nil {
		return err
	}

	query := `SELECT id, OrderID, uasID, start_time, end_time, start_lat, start_lng, end_lat, end_lng, distance, battery_used, created_at, payload, expressCount FROM flight_records WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND OrderID=?"
		args = append(args, orderID)
	}
	if uasID != "" {
		query += " AND uasID=?"
		args = append(args, uasID)
	}
	if startTime != "" {
		query += " AND start_time >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND end_time <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY start_time DESC"

	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			id, payload, expressCount          int
			orderIDs, uasIDs                   string
			startTimeT, endTimeT, createdAt    sql.NullTime
			startLat, startLng, endLat, endLng sql.NullInt64
			distance, batteryUsed              sql.NullFloat64
		)
		if err := rows.Scan(&id, &orderIDs, &uasIDs, &startTimeT, &endTimeT, &startLat, &startLng, &endLat, &endLng, &distance, &batteryUsed, &createdAt, &payload, &expressCount); err != nil {
			continue
		}
		var startLatVal string
		if startLat.Valid {
			startLatVal = strconv.FormatFloat(float64(startLat.Int64)/1e7, 'f', -1, 64)
		} else {
			startLatVal = ""
		}
		var startLngVal string
		if startLng.Valid {
			startLngVal = strconv.FormatFloat(float64(startLng.Int64)/1e7, 'f', -1, 64)
		} else {
			startLngVal = ""
		}
		var endLatVal string
		if endLat.Valid {
			endLatVal = strconv.FormatFloat(float64(endLat.Int64)/1e7, 'f', -1, 64)
		} else {
			endLatVal = ""
		}
		var endLngVal string
		if endLng.Valid {
			endLngVal = strconv.FormatFloat(float64(endLng.Int64)/1e7, 'f', -1, 64)
		} else {
			endLngVal = ""
		}
		payloadVal := strconv.FormatFloat(float64(payload)/10.0, 'f', -1, 64)

		row := []string{
			strconv.Itoa(id),
			orderIDs,
			uasIDs,
			nullableTimeToString(startTimeT),
			nullableTimeToString(endTimeT),
			startLatVal,
			startLngVal,
			endLatVal,
			endLngVal,
			nullableFloatToString(distance),
			nullableFloatToString(batteryUsed),
			nullableTimeToString(createdAt),
			payloadVal,
			strconv.Itoa(expressCount),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// ExportTrackPointsToCSVStream 使用流式写入将 flight_track_points 导出为 csv 文件
func (d *MySQLDao) ExportTrackPointsToCSVStream(startTime, endTime, orderID, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	bw := bufio.NewWriter(f)
	defer bw.Flush()
	w := csv.NewWriter(bw)
	defer w.Flush()

	headers := []string{"ID", "Order ID", "Flight Status", "Time Stamp", "Longitude", "Latitude", "Height Type", "Height", "Altitude", "VS (m/s)", "GS (m/s)", "Course", "SOC", "RM", "Voltage", "Current", "Wind Speed", "Wind Direct", "Temperature", "Humidity"}
	if err := w.Write(headers); err != nil {
		return err
	}

	query := `SELECT id, orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, voltage, current, windSpeed, windDirect, temperture, humidity FROM flight_track_points WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND orderID = ?"
		args = append(args, orderID)
	}
	if startTime != "" {
		query += " AND timeStamp >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND timeStamp <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY timeStamp ASC"

	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id                                          int64
			orderIDOut, flightStatus                    string
			timeStamp                                   sql.NullTime
			longitude, latitude                         sql.NullInt64
			heightType, height, altitude                sql.NullInt64
			VS, GS, course, SOC, RM, voltage, current   sql.NullInt64
			windSpeed, windDirect, temperture, humidity sql.NullInt64
		)
		if err := rows.Scan(&id, &orderIDOut, &flightStatus, &timeStamp, &longitude, &latitude, &heightType, &height, &altitude, &VS, &GS, &course, &SOC, &RM, &voltage, &current, &windSpeed, &windDirect, &temperture, &humidity); err != nil {
			continue
		}
		var lonVal, latVal, vsVal, gsVal string
		if longitude.Valid {
			lonVal = strconv.FormatFloat(float64(longitude.Int64)/1e7, 'f', -1, 64)
		} else {
			lonVal = ""
		}
		if latitude.Valid {
			latVal = strconv.FormatFloat(float64(latitude.Int64)/1e7, 'f', -1, 64)
		} else {
			latVal = ""
		}
		if VS.Valid {
			vsVal = strconv.FormatFloat(float64(VS.Int64)/10.0, 'f', -1, 64)
		} else {
			vsVal = ""
		}
		if GS.Valid {
			gsVal = strconv.FormatFloat(float64(GS.Int64)/10.0, 'f', -1, 64)
		} else {
			gsVal = ""
		}

		row := []string{
			strconv.FormatInt(id, 10),
			orderIDOut,
			flightStatus,
			nullableTimeToString(timeStamp),
			lonVal,
			latVal,
			nullableInt64ToString(heightType),
			nullableInt64ToString(height),
			nullableInt64ToString(altitude),
			vsVal,
			gsVal,
			nullableInt64ToString(course),
			nullableInt64ToString(SOC),
			nullableInt64ToString(RM),
			nullableInt64ToString(voltage),
			nullableInt64ToString(current),
			nullableInt64ToString(windSpeed),
			nullableInt64ToString(windDirect),
			nullableInt64ToString(temperture),
			nullableInt64ToString(humidity),
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	return nil
}

// helpers to convert nullable types used above to string
func nullableTimeToString(t sql.NullTime) string {
	if t.Valid {
		return t.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func nullableInt64ToString(n sql.NullInt64) string {
	if n.Valid {
		return strconv.FormatInt(n.Int64, 10)
	}
	return ""
}

func nullableFloatToString(n sql.NullFloat64) string {
	if n.Valid {
		return strconv.FormatFloat(n.Float64, 'f', -1, 64)
	}
	return ""
}

// helper formatting functions for nullable types
func nullableTimeFormat(t sql.NullTime) interface{} {
	if t.Valid {
		return t.Time.Format("2006-01-02 15:04:05")
	}
	return ""
}

func nullableInt64(n sql.NullInt64) interface{} {
	if n.Valid {
		return n.Int64
	}
	return ""
}

func nullableFloat64(n sql.NullFloat64) interface{} {
	if n.Valid {
		return n.Float64
	}
	return ""
}

// QueryTrackPoints 按时间范围或 orderID 查询轨迹点
func (d *MySQLDao) QueryTrackPoints(startTime, endTime, orderID string) ([]map[string]interface{}, error) {
	query := `SELECT id, orderID, flightStatus, timeStamp, longitude, latitude, heightType, height, altitude, VS, GS, course, SOC, RM, voltage, current, windSpeed, windDirect, temperture, humidity
		FROM flight_track_points WHERE 1=1`
	args := []interface{}{}
	if orderID != "" {
		query += " AND orderID = ?"
		args = append(args, orderID)
	}
	if startTime != "" {
		query += " AND timeStamp >= ?"
		args = append(args, startTime)
	}
	if endTime != "" {
		query += " AND timeStamp <= ?"
		args = append(args, endTime)
	}
	query += " ORDER BY timeStamp ASC"
	rows, err := d.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var points []map[string]interface{}
	for rows.Next() {
		var (
			id                                          int64
			orderIDOut, flightStatus                    string
			timeStamp                                   time.Time
			longitude, latitude                         int64
			heightType, height, altitude                int
			VS, GS, course, SOC, RM, voltage, current   int
			windSpeed, windDirect, temperture, humidity int
		)
		err := rows.Scan(&id, &orderIDOut, &flightStatus, &timeStamp, &longitude, &latitude, &heightType, &height, &altitude, &VS, &GS, &course, &SOC, &RM, &voltage, &current, &windSpeed, &windDirect, &temperture, &humidity)
		if err == nil {
			points = append(points, map[string]interface{}{
				"id":           id,
				"orderID":      orderIDOut,
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
				"voltage":      voltage,
				"current":      current,
				"windSpeed":    windSpeed,
				"windDirect":   windDirect,
				"temperture":   temperture,
				"humidity":     humidity,
			})
		}
	}
	return points, nil
}

// 更新指定架次的载货量
func (d *MySQLDao) UpdateFlightPayload(orderID string, payload, expressCount int) error {
	_, err := d.DB.Exec("UPDATE flight_records SET payload=?, expressCount=? WHERE OrderID=?", payload, expressCount, orderID)
	return err
}
