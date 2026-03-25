package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type NarodmonSensor struct {
	Id    string      `json:"id,omitempty"`
	Name  string      `json:"name,omitempty"`
	Value interface{} `json:"value,omitempty"`
	Unit  string      `json:"unit,omitempty"`
	Time  int         `json:"time,omitempty"`
}
type NarodmonDevice struct {
	Mac     string           `json:"mac,omitempty"`
	Name    string           `json:"name,omitempty"`
	Owner   string           `json:"owner,omitempty"`
	Lat     float64          `json:"lat,omitempty"`
	Lon     float64          `json:"lon,omitempty"`
	Alt     int              `json:"alt,omitempty"`
	Sensors []NarodmonSensor `json:"sensors,omitempty"`
}

type NarodmonJsonRequest struct {
	Devices []NarodmonDevice `json:"devices"`
}

type NarodmonHandler struct {
	client *http.Client
}

func NewNarodmonHandler() *NarodmonHandler {
	return &NarodmonHandler{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (h *NarodmonHandler) Name() string {
	return "narodmon"
}

func (h *NarodmonHandler) Handle(reading *Reading) error {
	payload := []NarodmonDevice{}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://narodmon.ru/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.Do(req)
	if err != nil {
		slog.Warn("narodmon handler: failed to send", "error", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	return fmt.Errorf("narodmon handler: server returned status %d", resp.StatusCode)
}
