package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func isHTTPS(url string) bool {
	return strings.HasPrefix(strings.ToLower(url), "https")
}

type DataCakeHandler struct {
	client *http.Client
}

type DataCakePayload map[string]any

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

func (h *DataCakeHandler) getClient(url string, cfg *HandlerConfig) *http.Client {
	skipVerify := false
	if cfg.DataCake.SkipVerify != nil {
		skipVerify = *cfg.DataCake.SkipVerify
	}
	if isHTTPS(url) && h.client.Transport == nil {
		return &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
			},
		}
	} else {

	}
	return h.client
}

func (h *DataCakeHandler) Handle(reading *Reading, cfg *HandlerConfig) error {

	if cfg.DataCake == nil || !cfg.DataCake.Enabled {
		return nil
	}
	payload := DataCakePayload{
		"device": reading.SensorMAC,
	}
	slog.Info("🎂 DataCake Handler Endpoint:", cfg.DataCake.Endpoint)
	payload[reading.Type] = reading.Value
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, cfg.DataCake.Endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := h.getClient(cfg.DataCake.Endpoint, cfg)
	resp, err := client.Do(req)
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
