package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"os/signal"

	"github.com/prokopparuzek/go-dht"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"libdb.so/hserve"
	"libdb.so/periph-gpioc/gpiodriver"
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
	dht     *dht.DHT
	metrics [2]*prometheus.Desc
}

func NewCollector(cfg *Config) (*Collector, error) {
	dht, err := dht.NewDHT(
		cfg.PinName,
		cfg.TemperatureUnit.toDHTConstant(),
		cfg.SensorType.toDHTConstant())
	if err != nil {
		return nil, fmt.Errorf("failed to create a new DHT11/DHT22 controller instance: %w", err)
	}

	metrics := [2]*prometheus.Desc{
		prometheus.NewDesc(
			"dht_temperature",
			"Temperature in the selected unit.",
			nil,
			joinMap(cfg.PrometheusLabels, map[string]string{
				"pin":         cfg.PinName,
				"sensor_type": string(cfg.SensorType),
				"unit":        string(cfg.TemperatureUnit),
			}),
		),
		prometheus.NewDesc(
			"dht_humidity",
			"Humidity percentage.",
			nil,
			joinMap(cfg.PrometheusLabels, map[string]string{
				"pin":         cfg.PinName,
				"sensor_type": string(cfg.SensorType),
			}),
		),
	}

	return &Collector{
		dht:     dht,
		metrics: metrics,
	}, nil
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	t, h, err := c.dht.Read()
	if err != nil {
		slog.Error(
			"failed to read the DHT11/DHT22 sensor",
			"err", err)
		return
	}

	slog.Info(
		"collected DHT11/DHT22 data",
		"temperature", t,
		"humidity", h)

	ch <- prometheus.MustNewConstMetric(c.metrics[0], prometheus.GaugeValue, t)
	ch <- prometheus.MustNewConstMetric(c.metrics[1], prometheus.GaugeValue, h)
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
