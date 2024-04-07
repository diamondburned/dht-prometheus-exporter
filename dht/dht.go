package dht

import "fmt"

// SensorType is the type of temperature/humidity sensor being used on the GPIO
// pin.
type SensorType string

const (
	DHT11 SensorType = "DHT11"
	DHT22 SensorType = "DHT22"
)

// CelsiusTemperature is a temperature in Celsius.
type CelsiusTemperature float64

func (t CelsiusTemperature) String() string {
	return fmt.Sprintf("%.1f°C", t)
}

// Humidity is a humidity percentage from 0% to 100%.
type Humidity float64

func (h Humidity) String() string {
	return fmt.Sprintf("%.1f%%", h)
}
