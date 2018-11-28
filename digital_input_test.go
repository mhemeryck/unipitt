package unipitt

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// setup for creating a temporary digital input
func setup(folder string) (dir string, filename string, f *os.File, err error) {
	// Create temporary folder, only if it does not exist already
	dir = filepath.Join(os.TempDir(), folder)
	if _, pathErr := os.Stat(dir); pathErr != nil {
		err = os.Mkdir(dir, os.ModePerm)
		if err != nil {
			return
		}
	}
	// Create temporary path
	tmpfn := filepath.Join(dir, DiFilename)
	// Create temporary file handle
	f, err = os.Create(tmpfn)
	if err != nil {
		return
	}
	return
}

func TestNewDigitalInput(t *testing.T) {
	// Setup
	folder := "di_1_01"
	name := "di_1_01"
	dir, filename, f, err := setup(folder)
	defer os.RemoveAll(dir)   // clean up
	defer os.Remove(filename) // clean up
	defer f.Close()
	if err != nil {
		t.Fatalf("Got error creating temporary file system setup: %s\n", err)
	}

	// Create a new DigitalInputReader
	digitalInput, err := NewDigitalInputReader(dir, name)

	// Test
	if digitalInput.Name != name {
		t.Fatalf("Expected digital input name to be %s, got %s\n", name, digitalInput.Name)
	}
	if digitalInput.Path != dir {
		t.Fatalf("Expected digital input path to be %s, got %s\n", folder, digitalInput.Path)
	}
	if err != nil {
		t.Fatalf("Got error reading file handle %s\n", err)
	}
}

func TestClose(t *testing.T) {
	// Setup
	folder := "di_1_01"
	name := "di_1_01"
	dir, filename, f, err := setup(folder)
	defer os.RemoveAll(dir)   // clean up
	defer os.Remove(filename) // clean up
	defer f.Close()
	if err != nil {
		t.Fatalf("Got error creating temporary file system setup: %s\n", err)
	}
	digitalInput, err := NewDigitalInputReader(dir, name)
	if err != nil {
		t.Fail()
	}

	// Test close
	err = digitalInput.Close()
	if err != nil {
		t.Fatalf("Got error closing file handle %s\n", err)
	}
}

func TestUpdate(t *testing.T) {
	// Setup
	folder := "di_1_01"
	name := "di_1_01"
	dir, filename, f, err := setup(folder)
	defer os.RemoveAll(dir)   // clean up
	defer os.Remove(filename) // clean up
	defer f.Close()
	if err != nil {
		t.Fatalf("Got error creating temporary file system setup: %s\n", err)
	}
	digitalInput, err := NewDigitalInputReader(dir, name)
	if err != nil {
		t.Fail()
	}
	cases := []struct {
		Previous bool
		Contents string
		Expected bool
		HasEvent bool
	}{
		{
			Previous: false,
			Contents: "0\n",
			Expected: false,
			HasEvent: false,
		},
		{
			Previous: false,
			Contents: "1\n",
			Expected: true,
			HasEvent: true,
		},
		{
			Previous: true,
			Contents: "1\n",
			Expected: true,
			HasEvent: false,
		},
	}
	events := make(chan *DigitalInputReader)
	defer close(events)
	for _, testCase := range cases {
		f.Seek(0, 0)
		_, err := f.WriteString(testCase.Contents)
		if err != nil {
			t.Fail()
		}
		digitalInput.Value = testCase.Previous
		go digitalInput.Update(events)
		if testCase.HasEvent {
			<-events
		}
		if digitalInput.Value != testCase.Expected {
			// Check it's true
			t.Fatalf("Expected value %t, got %t\n", testCase.Expected, digitalInput.Value)
		}
	}
}

func TestPoll(t *testing.T) {
	// File / folder setup
	folder := "di_1_01"
	name := "di_1_01"
	dir, filename, f, err := setup(folder)
	defer os.RemoveAll(dir)   // clean up
	defer os.Remove(filename) // clean up
	defer f.Close()
	if err != nil {
		t.Fatalf("Got error creating temporary file system setup: %s\n", err)
	}
	digitalInput, err := NewDigitalInputReader(dir, name)
	if err != nil {
		t.Fail()
	}

	// Events
	events := make(chan *DigitalInputReader)
	defer close(events)

	// Setup for triggering an event
	digitalInput.Value = false
	f.Seek(0, 0)
	_, err = f.WriteString("1\n")
	if err != nil {
		t.Fail()
	}

	// Poll
	go digitalInput.Poll(events, 500)

	// Block on events
	d := <-events

	// Check updated value
	if d.Value != true {
		t.Fatalf("Expected digital input value to be updated to %t, got %t", true, false)
	}
}

func TestPollError(t *testing.T) {
	folder := "di_1_01"
	name := "di_1_01"
	dir, filename, f, err := setup(folder)
	defer os.RemoveAll(dir)   // clean up
	defer os.Remove(filename) // clean up
	defer f.Close()
	if err != nil {
		t.Fatalf("Got error creating temporary file system setup: %s\n", err)
	}
	digitalInput, err := NewDigitalInputReader(dir, name)
	if err != nil {
		t.Fail()
	}

	// Events
	events := make(chan *DigitalInputReader)
	defer close(events)

	f.Close()
	// Poll
	go digitalInput.Poll(events, 500)

	d := <-events

	if d.Err == nil {
		t.Fatal("Expected an error on the returned digital input, found none")
	}
}

func TestFindDigitalInputReaders(t *testing.T) {
	folder := "di_1_01"
	// Create temporary folder, only if it does not exist already
	root, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, folder)
	if _, pathErr := os.Stat(dir); pathErr != nil {
		err := os.Mkdir(dir, os.ModePerm)
		if err != nil {
			t.Fail()
		}
	}
	defer os.RemoveAll(root) // clean up

	readers, err := FindDigitalInputReaders(root)

	if err != nil {
		t.Fatal(err)
	}
	if len(readers) != 1 {
		t.Fatalf("Expected 1 reader to be found, found %d\n", len(readers))
	}
	if readers[0].Name != folder {
		t.Fatalf("Expected name to be %s, found %s\n", folder, readers[0].Name)
	}
}

func TestFindDigitalInputReadersNonExisting(t *testing.T) {
	_, err := FindDigitalInputReaders("foo")
	if err == nil {
		t.Fatal(err)
	}
}
