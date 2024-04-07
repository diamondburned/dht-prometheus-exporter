package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/prokopparuzek/go-dht"
	"github.com/prometheus/client_golang/prometheus"
)

type Config struct {
	ListenAddr       string            `json:"listen_addr"`
	GPIOPin          int               `json:"gpio_pin"`
	SensorType       SensorType        `json:"sensor_type"`
	TemperatureUnit  TemperatureUnit   `json:"temperature_unit"`
	PrometheusLabels prometheus.Labels `json:"prometheus_labels,omitempty"`
}

func (c Config) Validate() error {
	switch c.SensorType {
	case DHT11, DHT22:
	default:
		return fmt.Errorf("unsupported sensor type: %q", c.SensorType)
	}

	switch c.TemperatureUnit {
	case Celsius, Fahrenheit:
	default:
		return fmt.Errorf("unsupported temperature unit: %q", c.TemperatureUnit)
	}

	return nil
}

// SensorType is the type of temperature/humidity sensor being used on the GPIO
// pin.
type SensorType string

const (
	DHT11 SensorType = "DHT11"
	DHT22 SensorType = "DHT22"
)

func (t SensorType) toDHTConstant() string {
	switch t {
	case DHT11:
		return "dht11"
	case DHT22:
		return "dht22"
	default:
		panic(fmt.Sprintf("unknown sensor type: %q", t))
	}
}

// TemperatureUnit is the unit of temperature to use.
type TemperatureUnit string

const (
	Celsius    TemperatureUnit = "celsius"
	Fahrenheit TemperatureUnit = "fahrenheit"
)

func (t TemperatureUnit) toDHTConstant() dht.TemperatureUnit {
	switch t {
	case Celsius:
		return dht.Celsius
	case Fahrenheit:
		return dht.Fahrenheit
	default:
		panic(fmt.Sprintf("unknown temperature unit: %q", t))
	}
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
