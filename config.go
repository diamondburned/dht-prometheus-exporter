package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"libdb.so/dht-prometheus-exporter/dht"
)

type Config struct {
	ListenAddr       string            `json:"listen_addr"`
	PinName          string            `json:"pin_name"`
	SensorType       dht.SensorType    `json:"sensor_type"`
	PrometheusLabels prometheus.Labels `json:"prometheus_labels,omitempty"`
}

func (c Config) Validate() error {
	switch c.SensorType {
	case dht.DHT11, dht.DHT22:
	default:
		return fmt.Errorf("unsupported sensor type: %q", c.SensorType)
	}
	return nil
}

func parseConfigFile(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &cfg, nil
}
