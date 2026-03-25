package scanner

import (
	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"
	"time"
)

func parseRaw(data []byte, obj *config.BLEObjectConfig, t time.Time) []handler.Reading {
	if len(data) < 1 {
		return nil
	}
	return []handler.Reading{
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "raw",
			Value:      float64(data[0]),
			Unit:       "",
			Timestamp:  t,
			Data:       data,
		},
	}
}
