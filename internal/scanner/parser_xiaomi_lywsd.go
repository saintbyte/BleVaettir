package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
	"github.com/saintbyte/BleVaettir/internal/handler"
)

type XiomiLywsdoSensorData struct {
	Temperature float64
	Humidity    float64
	Voltage     float64
	Battery     int
}

func XiomiLywsdo3mmcReadSensorData(mac string, timeout time.Duration, dev *linux.Device) (XiomiLywsdoSensorData, error) {
	var result XiomiLywsdoSensorData

	ble.SetDefaultDevice(dev)

	var (
		serviceUUID  = ble.MustParse("ebe0ccb0-7a0a-4b0c-8a1a-6ff2997da3a6")
		tempCharUUID = ble.MustParse("ebe0ccc1-7a0a-4b0c-8a1a-6ff2997da3a6")
		batCharUUID  = ble.MustParse("ebe0ccd8-7a0a-4b0c-8a1a-6ff2997da3a6")
	)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cln, err := ble.Connect(ctx, func(a ble.Advertisement) bool {
		return strings.EqualFold(a.Addr().String(), mac)
	})
	if err != nil {
		return result, fmt.Errorf("failed to connect: %v", err)
	}
	defer cln.CancelConnection()

	srvs, err := cln.DiscoverServices([]ble.UUID{serviceUUID})
	if err != nil {
		return result, fmt.Errorf("failed to discover services: %v", err)
	}
	if len(srvs) == 0 {
		return result, fmt.Errorf("service not found")
	}

	chars, err := cln.DiscoverCharacteristics(nil, srvs[0])
	if err != nil {
		return result, fmt.Errorf("failed to discover characteristics: %v", err)
	}

	var tempChar *ble.Characteristic
	var batChar *ble.Characteristic
	for _, c := range chars {
		if c.UUID.Equal(tempCharUUID) {
			tempChar = c
		}
		if c.UUID.Equal(batCharUUID) {
			batChar = c
		}
	}
	if tempChar == nil {
		return result, fmt.Errorf("temperature characteristic not found")
	}

	descs, err := cln.DiscoverDescriptors(nil, tempChar)
	if err != nil {
		return result, fmt.Errorf("failed to discover descriptors: %v", err)
	}

	cccdUUID := ble.MustParse("00002902-0000-1000-8000-00805f9b34fb")
	for _, d := range descs {
		if d.UUID.Equal(cccdUUID) {
			if err := cln.WriteDescriptor(d, []byte{0x01, 0x00}); err != nil {
				return result, fmt.Errorf("failed to write CCCD: %v", err)
			}
			break
		}
	}

	if batChar != nil {
		cln.WriteCharacteristic(batChar, []byte{0xf4, 0x01, 0x00}, true)
	}

	var (
		once sync.Once
		done = make(chan struct{})
	)

	if err := cln.Subscribe(tempChar, false, func(b []byte) {
		once.Do(func() {
			defer close(done)
			if len(b) >= 5 {
				sign := b[1] & (1 << 7)
				temp := (int(b[1]&0x7F)<<8 | int(b[0]))
				if sign != 0 {
					temp -= 32767
				}

				voltage := float64(uint16(b[3])|uint16(b[4])<<8) / 1000.0
				battery := int((voltage - 2.1) * 100)
				if battery > 100 {
					battery = 100
				}
				if battery < 0 {
					battery = 0
				}

				result = XiomiLywsdoSensorData{
					Temperature: float64(temp) / 100.0,
					Humidity:    float64(b[2]),
					Voltage:     voltage,
					Battery:     battery,
				}
			}
		})
	}); err != nil {
		return result, fmt.Errorf("failed to subscribe: %v", err)
	}

	select {
	case <-done:
		break
	case <-ctx.Done():
		cln.Unsubscribe(tempChar, false)
		return result, fmt.Errorf("timeout waiting for sensor data")
	}
	cln.Unsubscribe(tempChar, false)
	cln.CancelConnection()
	return result, nil
}

func parseXiaomiLYWSD(s *Scanner, a ble.Advertisement) []handler.Reading {
	obj := s.objectMap[a.Addr().String()]
	if obj == nil {
		return nil
	}

	result, err := XiomiLywsdo3mmcReadSensorData(obj.MAC, 30*time.Second, s.Device())
	if err != nil {
		slog.Error("failed to read XiomiLywsdo3mmc sensor data: %v", err)
	}

	t := time.Now()

	return []handler.Reading{
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "temperature",
			Value:      result.Temperature,
			Unit:       "°C", Timestamp: t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "humidity",
			Value:      result.Humidity,
			Unit:       "%",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "battery",
			Value:      float64(result.Battery),
			Unit:       "%",
			Timestamp:  t,
		},
		{
			SensorMAC:  obj.MAC,
			SensorName: obj.Name,
			Type:       "voltage",
			Value:      result.Voltage,
			Unit:       "V",
			Timestamp:  t,
		},
	}
}
