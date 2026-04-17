package scanner

import (
	"time"

	"github.com/go-ble/ble"
	"github.com/saintbyte/BleVaettir/internal/handler"
)

func parseRaw(s *Scanner, a ble.Advertisement) []handler.Reading {
	data := a.ManufacturerData()
	obj := s.objectMap[a.Addr().String()]
	if obj == nil {
		return nil
	}
	t := time.Now()
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
