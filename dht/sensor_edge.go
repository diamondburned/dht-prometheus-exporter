package dht

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/gpio"
)

type sensorEdgeDetection Sensor

func (s *sensorEdgeDetection) Read() ([]byte, error) {
	if err := s.start(); err != nil {
		return nil, fmt.Errorf("failed to start communication: %w", err)
	}
	if err := s.readAcknowledge(); err != nil {
		return nil, fmt.Errorf("failed to read acknowledge: %w", err)
	}
	data := make([]byte, 5)
	if _, err := s.readData(data); err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}
	if err := s.finish(); err != nil {
		return nil, fmt.Errorf("failed to finish communication: %w", err)
	}
	return data, nil
}

func (s *sensorEdgeDetection) start() error {
	// Set the pin to low for 20 milliseconds.
	if err := s.pin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set pin to low: %w", err)
	}

	time.Sleep(18 * time.Millisecond)

	// Pull the pin high and begin to read.
	if err := s.pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
		return fmt.Errorf("failed to set pin to input: %w", err)
	}

	// Wait until DHT responds.
	if !s.readFor(gpio.Low, 40*time.Microsecond) {
		return ErrTimeout
	}

	return nil
}

func (s *sensorEdgeDetection) readAcknowledge() error {
	// Wait for a high pull (after 80us) then a low pull (after 80us).
	ok := true &&
		s.readFor(gpio.High, 80*time.Microsecond) &&
		s.readFor(gpio.Low, 80*time.Microsecond)
	if !ok {
		return ErrTimeout
	}
	return nil
}

func (s *sensorEdgeDetection) readData(b []byte) (int, error) {
	const bit0Duration = 24 * time.Microsecond
	const bit1Duration = 70 * time.Microsecond

	for i := range b {
		var v byte
		for j := 0; j < 8; j++ {
			// Wait about 50us for DHT to pull high with our bit.
			if !s.readFor(gpio.High, 50*time.Microsecond) {
				return i, ErrTimeout
			}

			// DHT will pull high for either 24us (bit 0) or 70us (bit 1).
			// We wait for about 45us for it to pull back low.
			if s.readFor(gpio.Low, bit1Duration/2) {
				// We caught DHT pulling low early, so it's a 0 bit.
				v = v<<1 | 0b0
			} else {
				// DHT still hasn't pulled low, so it's a 1 bit.
				v = v<<1 | 0b1

				// Finish waiting for the 1 bit.
				if !s.readFor(gpio.Low, bit1Duration/2) {
					return i, ErrTimeout
				}
			}
		}
		b[i] = v
	}

	return len(b), nil
}

func (s *sensorEdgeDetection) finish() error {
	// Finish the line by ensuring it's pulled high.
	if !s.readFor(gpio.High, 80*time.Microsecond) {
		return ErrTimeout
	}
	return nil
}

// readFor reads the pin until the given level is reached or the timeout
// occurs.
func (s *sensorEdgeDetection) readFor(level gpio.Level, timeout time.Duration) bool {
	return s.pin.Read() == level || (s.pin.WaitForEdge(timeout) && s.pin.Read() == level)
}
