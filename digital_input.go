package unipitt

import (
	"bytes"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"
)

const (
	// Filename to check for
	Filename = "di_value"
	// TrueValue is the value considered to be true
	TrueValue = "1"
	// FolderRegex represents to regular expression used for finding the required file to read from
	FolderRegex = "di_[0-9]_[0-9]{2}"
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
func (d *DigitalInputReader) Poll(events chan *DigitalInputReader, interval int) {
	ticker := time.NewTicker(time.Duration(interval) * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ticker.C:
			err := d.Update(events)
			if err != nil {
				d.Err = err
				events <- d
				log.Printf("Error polling digital input with topic %s\n", d.Topic)
				return
			}
			if count%100 == 0 {
				count = 0
				log.Printf("Polling digital input %s ...\n", d.Topic)
			}
			count++
		}
	}
}

// Close closes the current open file handle
func (d *DigitalInputReader) Close() error {
	return d.f.Close()
}

// NewDigitalInputReader creates a new DigitalInput and opens the file handle
func NewDigitalInputReader(folder string, topic string) (d *DigitalInputReader, err error) {
	f, err := os.Open(path.Join(folder, Filename))
	d = &DigitalInputReader{Topic: topic, Path: folder, f: f}
	return
}

// findDigitalInputPaths finds all digital inputs in a given root folder
func findDigitalInputPaths(root string) (paths []string, err error) {
	// Compile regex first
	regex, err := regexp.Compile(FolderRegex)
	// Walk the folder structure
	err = filepath.Walk(root,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if regex.MatchString(info.Name()) {
				paths = append(paths, path)
			}
			return err
		})
	return
}

// FindDigitalInputReaders crawls the root (sys) folder to find any matching digial inputs and creates corresponding DigitalInputReader instances from these.
func FindDigitalInputReaders(root string) (readers []DigitalInputReader, err error) {
	// Find the paths first
	paths, err := findDigitalInputPaths(root)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Found %d matching digital input paths\n", len(paths))
	readers = make([]DigitalInputReader, len(paths))
	for k, folder := range paths {
		// Read topic as the trailing folder path
		_, topic := path.Split(folder)
		digitalInputReader, err := NewDigitalInputReader(folder, topic)
		if err != nil {
			log.Print(err)
		}
		readers[k] = *digitalInputReader
	}
	return
}
