package main

import (
	"context"
	"fmt"
	"log"
	"maps"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"libdb.so/dht-prometheus-exporter/dht"
	"libdb.so/hserve"
	"libdb.so/periph-gpioc/gpiodriver"
	"periph.io/x/conn/v3/gpio/gpioreg"
)

var (
	configFile = "config.json"
)

func main() {
	log.SetFlags(0)

	pflag.StringVarP(&configFile, "config", "c", configFile, "Path to the configuration file.")
	pflag.Parse()

	cfg, err := parseConfigFile(configFile)
	if err != nil {
		log.Fatalln("Failed to parse the configuration file:", err)
	}

	if err := gpiodriver.Register(); err != nil {
		log.Fatalln("Failed to register the GPIO driver:", err)
	}

	collector, err := NewCollector(cfg)
	if err != nil {
		log.Fatalln("Failed to create a new collector instance:", err)
	}

	collectors := prometheus.NewRegistry()
	collectors.MustRegister(collector)

	r := http.NewServeMux()
	r.Handle("/metrics", promhttp.HandlerFor(collectors, promhttp.HandlerOpts{Registry: collectors}))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := hserve.ListenAndServe(ctx, cfg.ListenAddr, r); err != nil {
		log.Fatalln("Failed to start the HTTP server:", err)
	}
}

// Collector is the DHT11/22 data collector for Prometheus.
type Collector struct {
	sensor  *dht.Sensor
	metrics [2]*prometheus.Desc
}

func NewCollector(cfg *Config) (*Collector, error) {
	pin := gpioreg.ByName(cfg.PinName)
	if pin == nil {
		return nil, fmt.Errorf("failed to find the GPIO pin: %q", cfg.PinName)
	}

	sensor, err := dht.NewSensor(pin, cfg.SensorType)
	if err != nil {
		return nil, fmt.Errorf("failed to create a new DHT11/DHT22 controller instance: %w", err)
	}

	labels := joinMap(cfg.PrometheusLabels, map[string]string{
		"pin":         cfg.PinName,
		"sensor_type": string(cfg.SensorType),
	})

	metrics := [2]*prometheus.Desc{
		prometheus.NewDesc(
			"dht_temperature",
			"Temperature in celsius.",
			nil,
			labels,
		),
		prometheus.NewDesc(
			"dht_humidity",
			"Humidity percentage.",
			nil,
			labels,
		),
	}

	return &Collector{
		sensor:  sensor,
		metrics: metrics,
	}, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	t, h, err := c.sensor.Read(context.Background())
	if err != nil {
		log.Println("Failed to read the DHT11/DHT22 sensor:", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(c.metrics[0], prometheus.GaugeValue, float64(t))
	ch <- prometheus.MustNewConstMetric(c.metrics[1], prometheus.GaugeValue, float64(h))
}

func joinMap[K comparable, V any](values ...map[K]V) map[K]V {
	result := maps.Clone(values[0])
	for _, m := range values[1:] {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
