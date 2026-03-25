package scanner

import (
	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"
	"time"
)

func parseXiaomiLYWSD(data []byte, obj *config.BLEObjectConfig, t time.Time) []handler.Reading {
	if len(data) < 6 || data[0] != 0x58 || data[1] != 0x4C {
		return nil
	}
	return []handler.Reading{
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "temperature",
			Value:      float64(int16(data[4])<<8|int16(data[3])) / 100.0,
			Unit:       "°C", Timestamp: t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "humidity",
			Value:      float64(data[5]),
			Unit:       "%",
			Timestamp:  t,
		},
	}
}
