package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	BLE            BLEConfig            `yaml:"ble"`
	Storage        StorageConfig        `yaml:"storage"`
	Handlers       []HandlerConfig      `yaml:"handlers"`
	BLEObjects     []BLEObjectConfig    `yaml:"ble_objects"`
	UnknownObjects UnknownObjectsConfig `yaml:"unknown_objects"`
	Intervals      IntervalsConfig      `yaml:"intervals"`
	Log            LogConfig            `yaml:"log"`
}

type BLEConfig struct {
	HCI int `yaml:"hci"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type StorageConfig struct {
	Path string `yaml:"path"`
}

type IntervalsConfig struct {
	ScanDurationSec int `yaml:"scan_duration_sec"`
	ScanIntervalSec int `yaml:"scan_interval_sec"`
}

type BLEObjectConfig struct {
	Name     string          `yaml:"name"`
	MAC      string          `yaml:"mac"`
	Parsers  []ParserConfig  `yaml:"parsers"`
	Handlers []ObjectHandler `yaml:"handlers,omitempty"`
}

type UnknownObjectsConfig struct {
	Enabled  bool            `yaml:"enabled"`
	Handlers []ObjectHandler `yaml:"handlers"`
}

type ParserConfig struct {
	Type string `yaml:"type"`
}

type ObjectHandler struct {
	Type string `yaml:"type"`
}

type HandlerConfig struct {
	Type     string                 `yaml:"type"`
	DB       *DBHandlerConfig       `yaml:"db,omitempty"`
	HTTP     *HTTPHandlerConfig     `yaml:"http,omitempty"`
	Narodmon *NarodmonHandlerConfig `yaml:"narodmon,omitempty"`
}

type NarodmonHandlerConfig struct {
	Enabled bool   `yaml:"enabled"`
	Owner   string `yaml:"owner"`
	Lat     string `yaml:"lat"`
	Lon     string `yaml:"lon"`
	Alt     string `yaml:"alt"`
}

type DBHandlerConfig struct {
	Enabled bool `yaml:"enabled"`
}

type HTTPHandlerConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
