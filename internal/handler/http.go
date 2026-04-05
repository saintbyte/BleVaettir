package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type HTTPHandler struct {
	client *http.Client
}

type HTTPPayload struct {
	SensorMAC  string  `json:"sensor_mac"`
	SensorName string  `json:"sensor_name"`
	Type       string  `json:"type"`
	Value      float64 `json:"value"`
	Unit       string  `json:"unit"`
	Timestamp  string  `json:"timestamp"`
}

func NewHTTPHandler() *HTTPHandler {
	return &HTTPHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *HTTPHandler) Name() string {
	return "http"
}

func (h *HTTPHandler) Handle(reading *Reading, cfg *HandlerConfig) error {
	if cfg.HTTP == nil || !cfg.HTTP.Enabled {
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

	req, err := http.NewRequest(http.MethodPost, cfg.HTTP.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if cfg.HTTP.APIKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.HTTP.APIKey))
	}

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("http handler: failed to send", "error", err, "endpoint", cfg.HTTP.Endpoint)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("http handler: server returned status %d", resp.StatusCode)
}
