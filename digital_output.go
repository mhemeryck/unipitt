package unipitt

import (
	"log"
	"os"
	"path"
)

const (
	// DoFilename to check for
	DoFilename = "do_value"
	// DoTrueValue digital output true value to write
	DoTrueValue = "1\n"
	// DoFalseValue digital output false value to write
	DoFalseValue = "0\n"
	// DoFolderRegex regular expression used for finding folders which contain digital output
	DoFolderRegex = "do_[0-9]_[0-9]{2}"
)

// DigitalOutput represents the digital outputs of the unipi board
type DigitalOutput interface {
	Update(bool) error
}

// DigitalOutputWriter implements the digital output specifically for writing outputs to files
type DigitalOutputWriter struct {
	Topic string
	Path  string
}

// Update writes the updated value to the digital output
func (d *DigitalOutputWriter) Update(value bool) (err error) {
	f, err := os.Create(path.Join(d.Path, DoFilename))
	defer f.Close()
	if err != nil {
		return err
	}

	if value {
		_, err = f.WriteString(DoTrueValue)
	} else {
		_, err = f.WriteString(DoFalseValue)
	}
	if err == nil {
		log.Printf("Update value of digital output %s to %t\n", d.Topic, value)
	}
	return err
}

// NewDigitalOutputWriter creates a new digital output writer instance from a a given matching folder
func NewDigitalOutputWriter(folder string) (d *DigitalOutputWriter) {
	// Read topic as the trailing folder path
	_, topic := path.Split(folder)
	return &DigitalOutputWriter{Topic: topic, Path: folder}
}

// FindDigitalOutputWriters generates the output writes from a given path
func FindDigitalOutputWriters(root string) (writerMap map[string]DigitalOutputWriter, err error) {
	paths, err := findPathsByRegex(root, DoFolderRegex)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Found %d matching digital output paths\n", len(paths))
	writerMap = make(map[string]DigitalOutputWriter)
	var d *DigitalOutputWriter
	for _, path := range paths {
		d = NewDigitalOutputWriter(path)
		writerMap[d.Topic] = *d
	}
	return
}
