package handler

import (
	"log/slog"
)

type LogHandler struct{}

func NewLogHandler() *LogHandler {
	return &LogHandler{}
}

func (h *LogHandler) Name() string {
	return "log"
}

func (h *LogHandler) Handle(reading *Reading, cfg *HandlerConfig) error {
	slog.Info("sensor reading",
		"mac", reading.SensorMAC,
		"name", reading.SensorName,
		"type", reading.Type,
		"value", reading.Value,
		"unit", reading.Unit,
		"handle", h.Name(),
		"data", string(reading.Data[:]),
	)
	return nil
}
