package flight

import "math"

// 地球半径（单位：米）
const EarthRadius = 6371000

// 计算新的经纬度
func GetNewLatLon(lat, lon, speed float64, bearing, interval int) (float64, float64) {
	// 角度转换为弧度
	latRad := degToRad(lat)
	lonRad := degToRad(lon)
	bearingRad := degToRad(float64(bearing))

	// 计算移动的距离（单位：米）
	distance := speed * float64(interval)

	// 计算新的纬度
	newLatRad := math.Asin(math.Sin(latRad)*math.Cos(distance/EarthRadius) +
		math.Cos(latRad)*math.Sin(distance/EarthRadius)*math.Cos(bearingRad))

	// 计算新的经度
	newLonRad := lonRad + math.Atan2(math.Sin(bearingRad)*math.Sin(distance/EarthRadius)*math.Cos(latRad),
		math.Cos(distance/EarthRadius)-math.Sin(latRad)*math.Sin(newLatRad))

	// 转回角度制
	return radToDeg(newLatRad), radToDeg(newLonRad)
}

// 角度转弧度
func degToRad(deg float64) float64 {
	return deg * math.Pi / 180
}

// 弧度转角度
func radToDeg(rad float64) float64 {
	return rad * 180 / math.Pi
}
