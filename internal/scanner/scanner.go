package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

type Scanner struct {
	cfg             *config.Config
	objectHandlers  map[string][]handler.Handler
	unknownHandlers []handler.Handler
	objectMap       map[string]*config.BLEObjectConfig
	hciID           int
}

func New(cfg *config.Config, objectHandlers map[string][]handler.Handler, unknownHandlers []handler.Handler) (*Scanner, error) {
	dev, err := linux.NewDevice(ble.OptDeviceID(cfg.BLE.HCI))
	if err != nil {
		return nil, fmt.Errorf("failed to open HCI%d: %w", cfg.BLE.HCI, err)
	}
	ble.SetDefaultDevice(dev)

	sc := &Scanner{
		cfg:             cfg,
		objectHandlers:  objectHandlers,
		unknownHandlers: unknownHandlers,
		objectMap:       make(map[string]*config.BLEObjectConfig),
		hciID:           cfg.BLE.HCI,
	}

	for i := range cfg.BLEObjects {
		obj := &cfg.BLEObjects[i]
		sc.objectMap[obj.MAC] = obj
	}

	return sc, nil
}

func (s *Scanner) Run(stop <-chan struct{}) {
	slog.Info("BLE scanner started", "hci", s.hciID, "objects", len(s.cfg.BLEObjects))

	scanInterval := time.Duration(s.cfg.Intervals.ScanIntervalSec) * time.Second
	ticker := time.NewTicker(scanInterval)
	defer ticker.Stop()

	s.scan()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			s.scan()
		}
	}
}

func (s *Scanner) scan() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.Intervals.ScanIntervalSec)*time.Second)
	defer cancel()

	seen := make(map[string]bool)
	var mu sync.Mutex

	err := ble.Scan(ctx, false, func(a ble.Advertisement) {
		mac := a.Addr().String()

		mu.Lock()
		if seen[mac] {
			mu.Unlock()
			return
		}
		seen[mac] = true
		mu.Unlock()

		if obj, ok := s.objectMap[mac]; ok {
			s.handleObject(a, obj)
		} else if len(s.unknownHandlers) > 0 {
			s.handleUnknown(a, mac)
		}
	}, nil)
	if err != nil && ctx.Err() == nil {
		slog.Warn("BLE scan error", "error", err)
	}
}

func (s *Scanner) handleObject(a ble.Advertisement, obj *config.BLEObjectConfig) {
	readings := s.parseAdvertisement(a, obj)
	handlers := s.objectHandlers[obj.MAC]

	for _, r := range readings {
		for _, h := range handlers {
			if err := h.Handle(&r); err != nil {
				slog.Warn("handler failed",
					"handler", h.Name(),
					"sensor", obj.Name,
					"error", err,
				)
			}
		}
	}
}

func (s *Scanner) handleUnknown(a ble.Advertisement, mac string) {
	data := a.ManufacturerData()
	now := time.Now()

	r := handler.Reading{
		SensorMAC:  mac,
		SensorName: "unknown",
		Type:       "raw",
		Value:      float64(len(data)),
		Unit:       "bytes",
		Timestamp:  now,
	}

	for _, h := range s.unknownHandlers {
		if err := h.Handle(&r); err != nil {
			slog.Warn("unknown handler failed",
				"handler", h.Name(),
				"mac", mac,
				"error", err,
			)
		}
	}
}

func (s *Scanner) parseAdvertisement(a ble.Advertisement, obj *config.BLEObjectConfig) []handler.Reading {
	var readings []handler.Reading
	now := time.Now()
	data := a.ManufacturerData()

	for _, parser := range obj.Parsers {
		switch parser.Type {
		case "xiaomi_lywsd03mmc":
			readings = append(readings, parseXiaomiLYWSD(data, obj, now)...)
		case "atc_thermometer":
			readings = append(readings, parseATC(data, obj, now)...)
		case "raw":
			readings = append(readings, parseRaw(data, obj, now)...)
		default:
			slog.Warn("unknown parser type", "type", parser.Type)
		}
	}

	return readings
}

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

func (s *Scanner) Close() error {
	return ble.Stop()
}
