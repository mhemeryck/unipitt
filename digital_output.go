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
	DoFolderRegex = "[dr]o_[0-9]_[0-9]{2}"
)

// DigitalOutput represents the digital outputs of the unipi board
type DigitalOutput interface {
	Update(bool) error
}

// DigitalOutputWriter implements the digital output specifically for writing outputs to files
type DigitalOutputWriter struct {
	Name string
	Path string
}

// Update writes the updated value to the digital output
func (d *DigitalOutputWriter) Update(value bool) (err error) {
	var value_filename string;

	value_filename = DoFilename;
	value_filename = d.Name[0:1] + DoFilename[1:]
	
	f, err := os.Create(path.Join(d.Path, value_filename))
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
		log.Printf("Update value of digital output %s to %t in file %s\n", d.Name, value, value_filename)
	}
	return err
}

// NewDigitalOutputWriter creates a new digital output writer instance from a a given matching folder
func NewDigitalOutputWriter(folder string) (d *DigitalOutputWriter) {
	// Read name as the trailing folder path
	_, name := path.Split(folder)
	return &DigitalOutputWriter{Name: name, Path: folder}
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
		writerMap[d.Name] = *d
	}
	return
}
