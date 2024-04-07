# DHT11/DHT22 Prometheus Exporter

This Go project exposes temperature and humidity readings from a DHT11 or DHT22
sensor as Prometheus metrics. 

## How to Run

1. **Configure:** Create a `config.json` file (see sample below).
2. **Install:** `go install libdb.so/dht-prometheus-exporter@latest`
4. **Run:** `dht-prometheus-exporter -c config.json`
5. **Access metrics:** Visit `http://[listen_addr]/metrics` 

**Sample `config.json`:** 

```json
{
  "listen_addr": ":9090", 
  "pin_name": "GPIO4",
  "sensor_type": "DHT22",
  "temperature_unit": "celsius",
  "prometheus_labels": {
    "hostname": "bridget"
  }
}
```

Listen address can be in the following formats:

- `:9090`, all interfaces on port 9090
- `localhost:9090`, only localhost on port 9090
- `unix:///tmp/dht-exporter.sock`, UNIX socket at `/tmp/dht-exporter.sock`
