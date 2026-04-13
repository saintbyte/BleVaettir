package scanner

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/saintbyte/BleVaettir/internal/config"
	"github.com/saintbyte/BleVaettir/internal/handler"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux"
)

type Scanner struct {
	cfg             *config.Config
	objectHandlers  map[string][]HandlerWithConfig
	unknownHandlers []HandlerWithConfig
	objectMap       map[string]*config.BLEObjectConfig
	hciID           int
	scanResults     []scanResult
	scope           Scope
}

func New(cfg *config.Config, objectHandlers map[string][]HandlerWithConfig, unknownHandlers []HandlerWithConfig) (*Scanner, error) {
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
		scanResults:     make([]scanResult, 0, 100),
		scope:           make(map[string]DeviceScope, 5),
	}

	for i := range cfg.BLEObjects {
		obj := &cfg.BLEObjects[i]
		sc.objectMap[obj.MAC] = obj
	}

	return sc, nil
}

func (s *Scanner) Run(stop <-chan struct{}) {
	slog.Info("BLE scanner started", "hci", s.hciID, "objects", len(s.cfg.BLEObjects))
	// Время между сканировани: время сканирования + время между сканирования - потому что так понятнее
	scanInterval := time.Duration(
		s.cfg.Intervals.ScanIntervalSec+s.cfg.Intervals.ScanDurationSec,
	) * time.Second
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
	s.scanResults = s.scanResults[:0]

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.Intervals.ScanDurationSec)*time.Second)
	defer cancel()

	err := ble.Scan(ctx, false, func(a ble.Advertisement) {
		s.scanResults = append(s.scanResults, scanResult{
			mac: a.Addr().String(),
			adv: a,
		})
	}, nil)

	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
		case errors.Is(err, context.DeadlineExceeded):
		default:
			slog.Warn("BLE scan error", "error", err)
		}
	}
	s.Close()
	s.processResults()
}

func (s *Scanner) processResults() {
	slog.Debug("Processing scan results", "count", len(s.scanResults))

	seenUnknown := make(map[string]bool)
	for _, r := range s.scanResults {
		if obj, ok := s.objectMap[r.mac]; ok {
			s.handleObject(r.adv, obj)
		} else if !seenUnknown[r.mac] && len(s.unknownHandlers) > 0 {
			s.handleUnknown(r.adv, r.mac)
			seenUnknown[r.mac] = true
		}
	}
	s.AfterHandleAll()
}

func (s *Scanner) handleObject(
	a ble.Advertisement,
	obj *config.BLEObjectConfig,
) {
	readings := s.parseAdvertisement(a, obj)
	handlers := s.objectHandlers[obj.MAC]

	for _, r := range readings {
		for _, h := range handlers {
			if err := h.Handler.Handle(&r, h.Config); err != nil {
				slog.Warn("handler failed",
					"handler", h.Handler.Name(),
					"sensor", obj.Name,
					"error", err,
				)
			}
		}
	}
	s.AfterHandleAllInObject()
}

func (s *Scanner) AfterHandleAllInObject() {
	slog.Info("event processed  all in obj")
}

func (s *Scanner) AfterHandleAll() {
	slog.Info("event processed all available object")
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
		if err := h.Handler.Handle(&r, h.Config); err != nil {
			slog.Warn("unknown handler failed",
				"handler", h.Handler.Name(),
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
		case "jaalee":
			readings = append(readings, parseJaalee(data, obj, now)...)
		case "raw":
			readings = append(readings, parseRaw(data, obj, now)...)
		default:
			slog.Warn("unknown parser type", "type", parser.Type)
		}
	}

	return readings
}

func (s *Scanner) Close() error {
	return ble.Stop()
}
