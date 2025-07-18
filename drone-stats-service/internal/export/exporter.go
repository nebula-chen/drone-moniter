package export

import (
	"drone-stats-service/internal/types"
	"encoding/csv"
	"fmt"
	"os"
)

func ExportFlightRecordsToCSV(records []types.FlightRecord, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"OrderID", "StartTime", "EndTime", "Longitude", "Latitude", "Altitude", "Distance", "SOC"})
	for _, r := range records {
		writer.Write([]string{
			fmt.Sprintf("%d", r.ID),
			r.OrderID,
			r.UasID,
			r.StartTime,
			r.EndTime,
			fmt.Sprintf("%d", r.StartLng),
			fmt.Sprintf("%d", r.StartLat),
			fmt.Sprintf("%d", r.EndLng),
			fmt.Sprintf("%d", r.EndLat),
			fmt.Sprintf("%.2f", r.Distance),
			fmt.Sprintf("%.2f", r.BatteryUsed),
		})
	}
	return nil
}
