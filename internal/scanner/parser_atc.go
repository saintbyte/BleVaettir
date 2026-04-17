package scanner

import (
	"time"

	"github.com/go-ble/ble"
	"github.com/saintbyte/BleVaettir/internal/handler"
)

func parseATC(s *Scanner, a ble.Advertisement) []handler.Reading {
	data := a.ManufacturerData()
	obj := s.objectMap[a.Addr().String()]
	if obj == nil {
		return nil
	}
	t := time.Now()
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
