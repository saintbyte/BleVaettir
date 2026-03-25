package scanner

import (
	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"
	"math"
	"time"
)

func parseJaalee(data []byte, obj *config.BLEObjectConfig, t time.Time) []handler.Reading {
	if len(data) < 26 {
		return nil
	}
	batteryLevel := int(data[25])
	rssi_ := data[24]
	temperature_ := (uint16(data[20]) << 8) | uint16(data[21])
	humidity_ := (uint16(data[22]) << 8) | uint16(data[23])

	digits := 2
	multiplier := math.Pow10(digits)
	temperature := math.Round(((float64(temperature_)/65535.0)*175-45)*multiplier) / multiplier
	humidity := math.Round(((float64(humidity_)/65535.0)*100)*multiplier) / multiplier
	return []handler.Reading{
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "temperature",
			Value:      temperature,
			Unit:       "°C",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "humidity",
			Value:      humidity,
			Unit:       "%",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "battery",
			Value:      float64(batteryLevel),
			Unit:       "%",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "rssi",
			Value:      float64(int8(rssi_)),
			Unit:       "db",
			Timestamp:  t,
		},
	}

}
