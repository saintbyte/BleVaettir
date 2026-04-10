package handler

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
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

func (h *HTTPHandler) getClient(cfg *HTTPHandlerConfig) *http.Client {
	if cfg == nil || (cfg.CACert == "" && cfg.ClientCert == "" && !*cfg.SkipVerify) {
		return h.client
	}

	tlsConfig := &tls.Config{}

	if *cfg.SkipVerify {
		tlsConfig.InsecureSkipVerify = true
	}

	if cfg.CACert != "" {
		caCert, err := os.ReadFile(cfg.CACert)
		if err != nil {
			slog.Error("http handler: failed to read CA cert", "error", err, "path", cfg.CACert)
		} else {
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
	}

	if cfg.ClientCert != "" && cfg.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(cfg.ClientCert, cfg.ClientKey)
		if err != nil {
			slog.Error("http handler: failed to load client cert", "error", err)
		} else {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
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

	payload := map[string]any{
		"device": reading.SensorMAC,
	}
	payload[reading.Type] = reading.Value
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

	resp, err := h.getClient(cfg.HTTP).Do(req)
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
