package unipitt

import (
	"log"
	"os"
	"path"
	"time"
)

const (
	// DiFilename to check for
	DiFilename = "di_value"
	// DiTrueValue is the value considered to be true
	DiTrueValue = "1"
	// DiFolderRegex represents to regular expression used for finding the required file to read from
	DiFolderRegex = "di_[0-9]_[0-9]{2}"
)

// DigitalInput interface for doing the polling
type DigitalInput interface {
	Update(chan *DigitalInput) error
	Poll(chan *DigitalInput, int)
	Close()
}

// DigitalInputReader implements the digital input interface
type DigitalInputReader struct {
	Name  string
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
	value := string(b) == DiTrueValue
	// Push out an event in case of a leading edge
	if d.Value != value {
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
				log.Printf("Error polling digital input with name %s\n", d.Name)
				return
			}
			if count%100 == 0 {
				count = 0
				log.Printf("Polling digital input %s ...\n", d.Name)
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
func NewDigitalInputReader(folder string, name string) (d *DigitalInputReader, err error) {
	f, err := os.Open(path.Join(folder, DiFilename))
	d = &DigitalInputReader{Name: name, Path: folder, f: f}
	return
}

// FindDigitalInputReaders crawls the root (sys) folder to find any matching digial inputs and creates corresponding DigitalInputReader instances from these.
func FindDigitalInputReaders(root string) (readers []DigitalInputReader, err error) {
	// Find the paths first
	paths, err := findPathsByRegex(root, DiFolderRegex)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Found %d matching digital input paths\n", len(paths))
	readers = make([]DigitalInputReader, len(paths))
	for k, folder := range paths {
		// Read name as the trailing folder path
		_, name := path.Split(folder)
		digitalInputReader, err := NewDigitalInputReader(folder, name)
		if err != nil {
			log.Print(err)
		}
		readers[k] = *digitalInputReader
	}
	return
}
