package handler

import (
	"log/slog"
	"time"
)

type Reading struct {
	SensorMAC  string
	SensorName string
	Type       string
	Value      float64
	Unit       string
	Timestamp  time.Time
	Data       []byte
}

type Handler interface {
	Handle(reading *Reading) error
	Name() string
}

type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Name() string {
	return "log"
}

func (h *LogHandler) Handle(reading *Reading) error {
	slog.Info("sensor reading",
		"mac", reading.SensorMAC,
		"name", reading.SensorName,
		"type", reading.Type,
		"value", reading.Value,
		"unit", reading.Unit,
		"data", string(reading.Data[:]),
	)
	return nil
}
