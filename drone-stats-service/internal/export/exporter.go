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
	writer.Write([]string{"FlightCode", "StartTime", "EndTime", "Longitude", "Latitude", "Altitude", "Distance", "SOC"})
	for _, r := range records {
		writer.Write([]string{
			r.FlightCode,
			r.FlightStatus,
			r.TimeStamp,
			fmt.Sprintf("%d", r.Longitude),
			fmt.Sprintf("%d", r.Latitude),
			fmt.Sprintf("%d", r.Altitude),
			fmt.Sprintf("%d", r.SOC),
		})
	}
	return nil
}
