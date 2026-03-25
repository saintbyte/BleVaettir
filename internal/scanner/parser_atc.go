package scanner

import (
	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"
	"time"
)

func parseATC(data []byte, obj *config.BLEObjectConfig, t time.Time) []handler.Reading {
	if len(data) < 10 {
		return nil
	}
	return []handler.Reading{
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "temperature",
			Value:      float64(int16(data[8])<<8|int16(data[7])) / 100.0,
			Unit:       "°C",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "humidity",
			Value:      float64(data[9]),
			Unit:       "%",
			Timestamp:  t,
		},
	}
}
