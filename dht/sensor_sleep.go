package dht

// type sensorSleep Sensor
//
// func (s *sensorSleep) start() error {
// 	const startDelay = 20 * time.Millisecond
//
// 	// Set the pin to low for 20 milliseconds.
// 	if err := s.pin.Out(gpio.Low); err != nil {
// 		return fmt.Errorf("failed to set pin to low: %w", err)
// 	}
// 	time.Sleep(startDelay)
//
// 	// Pull the pin high and begin to read.
// 	if err := s.pin.In(gpio.PullUp, gpio.BothEdges); err != nil {
// 		return fmt.Errorf("failed to set pin to input: %w", err)
// 	}
//
// 	return nil
// }
//
// func (s *sensorSleep) readData() ([]byte, error) {
// 	data := make([]byte, 5)
//
// 	for i := range data {
// 		// Wait for our falling (low) edge.
// 		_, err := readFor(s.pin, gpio.Low)
// 		if err != nil {
// 			return data, fmt.Errorf("failed to read first falling edge: %w", err)
// 		}
// 	}
// }

// func readFor(pin gpio.PinIO, until gpio.Level) (gpio.Level, error) {
// 	const maxTries = 100
//
// 	for i := 0; i < maxTries; i++ {
// 		if v := pin.Read(); v != until {
// 			return v, nil
// 		}
// 	}
//
// 	return gpio.Low, ErrTimeout
// }
