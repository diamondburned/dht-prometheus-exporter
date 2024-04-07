package dht

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"runtime"
	"time"

	"periph.io/x/conn/v3/gpio"
)

var (
	// ErrTimeout is returned when reading data from the sensor times out.
	ErrTimeout = errors.New("timeout while reading data")
	// ErrInvalidChecksum is returned when the data checksum is invalid.
	ErrInvalidChecksum = errors.New("invalid checksum")
)

// Sensor represents a DHT11/DHT22 device. A sensor instance is technically safe
// to be used concurrently but is likely useless when doing so.
type Sensor struct {
	pin  gpio.PinIO
	wait *time.Timer
	typ  SensorType
}

// NewSensor creates a new DHT11/DHT22 sensor on the given GPIO pin.
func NewSensor(pin gpio.PinIO, typ SensorType) (*Sensor, error) {
	switch typ {
	case DHT11, DHT22:
	default:
		return nil, fmt.Errorf("unknown sensor type: %s", typ)
	}

	if err := pin.Out(gpio.High); err != nil {
		return nil, fmt.Errorf("failed to set pin to high: %w", err)
	}

	return &Sensor{
		pin:  pin,
		wait: time.NewTimer(0),
		typ:  typ,
	}, nil
}

// Read reads the temperature and humidity data from the sensor.
//
// Note that the DHT11 sensor can only be read once every second and the DHT22
// sensor once every two seconds. The function automatically waits for the
// minimum interval to pass before reading the data unless the context is
// canceled.
func (s *Sensor) Read(ctx context.Context) (CelsiusTemperature, Humidity, error) {
	select {
	case <-s.wait.C:
		defer s.wait.Reset(s.minReadInterval())
	case <-ctx.Done():
		return 0, 0, ctx.Err()
	}

	data, err := s.readBytes()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read data: %w", err)
	}

	t, h := parseData(data, s.typ)
	return t, h, nil
}

func (s *Sensor) readBytes() ([]byte, error) {
	// Reference:
	// https://github.com/efthymios-ks/AVR-DHT/blob/master/Files/DHT.c#L41
	// http://www.ocfreaks.com/basics-interfacing-dht11-dht22-humidity-temperature-sensor-mcu/

	// NOTE:
	//
	// We have 2 ways of reading data: either by sleeping until the middle
	// of a signal edge (sleep) and check for its value or use the edge
	// detection API (edge detection).
	//
	// periph.io's edge detection API is extremely limited and doesn't expose
	// the event timestamp, which is required to calculate how long the edge
	// lasted. This is important because the DHT11/DHT22 protocol uses the
	// duration of a signal edge to encode bit on/off.
	//
	// The sleeping method works fine as long as the sleep latency is less than
	// ~40 microseconds.
	//
	// We might be able to use the edge detection method with our own time.Since
	// calls to detect the duration of the edge, but this is not guaranteed to
	// work as expected.

	runtime.LockOSThread()
	b, err := (*sensorEdgeDetection)(s).Read()
	runtime.UnlockOSThread()

	if err != nil {
		return nil, err
	}

	if err := validateDataChecksum(b); err != nil {
		return b, err
	}

	return b, nil
}

func (s *Sensor) minReadInterval() time.Duration {
	if s.typ == DHT11 {
		return 1 * time.Second
	}
	return 2 * time.Second
}

func parseData(b []byte, typ SensorType) (CelsiusTemperature, Humidity) {
	_ = b[:5] // skip bounds check

	var h Humidity
	switch typ {
	case DHT11:
		h = Humidity(float64(b[0]))
	case DHT22:
		h = Humidity(float64(binary.BigEndian.Uint16(b[:2])) / 10)
	}

	// Temperature.
	var t CelsiusTemperature
	switch typ {
	case DHT11:
		t = CelsiusTemperature(float64(b[2]))
	case DHT22:
		valueU16 := binary.BigEndian.Uint16(b[2:])
		negative := -float64(valueU16 >> 15) // -1 if the highest bit is set.
		t = CelsiusTemperature(negative * float64(valueU16) / 10)
	}

	return t, h
}

func validateDataChecksum(data []byte) error {
	// https://github.com/efthymios-ks/AVR-DHT/blob/858c39d8c5febe556e68050653d837924c0b613d/Files/DHT.c#L156
	// What the fuck. Why does the C code do this.
	var checksum uint8
	for _, b := range data[:4] {
		checksum += b
	}

	if checksum != data[4] {
		return ErrInvalidChecksum
	}

	return nil
}
