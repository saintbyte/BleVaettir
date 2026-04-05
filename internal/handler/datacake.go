package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type DataCakeHandler struct {
	client *http.Client
}

type DataCakePayload struct {
	SensorMAC  string  `json:"sensor_mac"`
	SensorName string  `json:"sensor_name"`
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  string  `json:"timestamp"`
}

func NewDataCakeHandler() *DataCakeHandler {
	return &DataCakeHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *DataCakeHandler) Name() string {
	return "datacake"
}

func (h *DataCakeHandler) Handle(reading *Reading, cfg *HandlerConfig) error {
	if cfg.DataCake == nil || !cfg.DataCake.Enabled {
		return nil
	}

	payload := []HTTPPayload{
		{
			SensorMAC:  reading.SensorMAC,
			SensorName: reading.SensorName,
			Type:       reading.Type,
			Value:      reading.Value,
			Unit:       reading.Unit,
			Timestamp:  reading.Timestamp.Format(time.RFC3339),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, cfg.DataCake.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("datacake handler: failed to send", "error", err, "endpoint", cfg.DataCake.Endpoint)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("datacake handler: server returned status %d", resp.StatusCode)
}
