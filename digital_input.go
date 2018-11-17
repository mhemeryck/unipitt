package unipitt

import (
	"bytes"
	"log"
	"os"
	"path"
	"time"
)

const (
	// Filename to check for
	Filename = "di_value"
	// TrueValue is the value considered to be true
	TrueValue = "1"
)

// DigitalInput interface for doing the polling
type DigitalInput interface {
	Update(chan *DigitalInput) error
	Poll(chan *DigitalInput, int)
	Close()
}

// DigitalInputReader implements the digital input interface
type DigitalInputReader struct {
	Topic string
	Value bool
	Path  string
	Err   error
	f     *os.File
}

// NewDigitalInputReader creates a new DigitalInput and opens the file handle
func NewDigitalInputReader(folder string, topic string) (d *DigitalInputReader, err error) {
	f, err := os.Open(path.Join(folder, Filename))
	d = &DigitalInputReader{Topic: topic, Path: folder, f: f}
	return
}

// Update reads the value and sets the new value
func (d *DigitalInputReader) Update(events chan *DigitalInputReader) (err error) {
	// Read the first byte
	d.f.Seek(0, 0)
	b := make([]byte, 1)
	_, err = d.f.Read(b)
	// Check it's true
	value := bytes.Equal(b, []byte(TrueValue))
	// Push out an event in case of a leading edge
	if !d.Value && value {
		events <- d
	}
	// Update value
	d.Value = value
	return
}

// Poll continuously updates the instance
func (d *DigitalInputReader) Poll(events chan *DigitalInputReader, ticker *time.Ticker) {
	for {
		select {
		case <-ticker.C:
			err := d.Update(events)
			if err != nil {
				d.Err = err
				events <- d
				log.Println("Error polling the digital input")
				return
			}
		}
	}
}

// Close closes the current open file handle
func (d *DigitalInputReader) Close() error {
	return d.f.Close()
}
