package handler

import (
	"log/slog"

	"github.com/saintbyte/BleVaettir/internal/storage"
)

type DBHandler struct {
	store *storage.Storage
}

func NewDBHandler(store *storage.Storage) *DBHandler {
	return &DBHandler{store: store}
}

func (h *DBHandler) Name() string {
	return "db"
}

func (h *DBHandler) Handle(reading *Reading) error {
	r := &storage.Reading{
		SensorMAC:  reading.SensorMAC,
		SensorName: reading.SensorName,
		Type:       reading.Type,
		Value:      reading.Value,
		Unit:       reading.Unit,
		Timestamp:  reading.Timestamp,
	}
	if err := h.store.Save(r); err != nil {
		slog.Warn("db handler: failed to save", "error", err, "sensor", reading.SensorName)
		return err
	}
	return nil
}
