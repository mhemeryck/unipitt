package unipitt

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// End-to-end test of handler.
//
// A dummy MQTT client is set up as well as a digital input from some temporary files. A corresponding handler is created. Afterwards, the polling loop starts (blocking). We make a change to the digital input file and see it is captured by looking at the logs. Afterwards, the handler can quit by passing in a done signal.
func TestHandler(t *testing.T) {
	pollingInterval := 50

	// Setup a folder structure
	folder := "di_1_01"
	// Create temporary folder, only if it does not exist already
	root, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	sysFsRoot := filepath.Join(root, folder)
	if _, pathErr := os.Stat(sysFsRoot); pathErr != nil {
		err := os.Mkdir(sysFsRoot, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}
	// Create temporary path
	tmpfn := filepath.Join(sysFsRoot, "di_value")
	// Create temporary file handle
	f, err := os.Create(tmpfn)
	if err != nil {
		t.Fatal(err)
	}
	// Put in zero-value
	_, err = f.WriteString("0\n")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	defer os.RemoveAll(root) // clean up

	config := Configuration{
		SysFsRoot: sysFsRoot,
		MQTT: MQTTConfig{
			Broker:   "mqtts://foo",
			ClientID: "unipitt",
		},
	}
	handler, err := NewHandler(config)
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()

	// Setup log monitoring
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Start polling (blocking)
	done := make(chan bool)
	// Trigger a send
	go func() {
		// Trigger a send
		f.Seek(0, 0)
		_, err = f.WriteString("1\n")
		if err != nil {
			t.Fatal(err)
		}
		// Some ugly waiting until everything has settled ...
		time.Sleep(1 * time.Second)
		done <- true
	}()
	handler.Poll(done, pollingInterval)
	if !bytes.Contains(buf.Bytes(), []byte("Trigger for name di_1_01")) {
		t.Fatal("Expected a trigger to be captured in the log, found none")
	}
	if !bytes.Contains(buf.Bytes(), []byte("Error connecting to MQTT broker ...")) {
		t.Fatal("Expected a reconnect for MQTT broker, did not find one")
	}
}

func TestHandlerNoSysFsRoot(t *testing.T) {
	config := Configuration{
		SysFsRoot: "foo",
		MQTT: MQTTConfig{
			Broker:   "mqtts://foo",
			ClientID: "unipitt",
		},
	}

	handler, err := NewHandler(config)
	if err == nil {
		t.Fatalf("Expected an error without a valid sys fs folder, got none")
	}
	defer handler.Close()
}

func TestHandlerCaFileIssue(t *testing.T) {

	// Setup a folder structure
	folder := "di_1_01"
	// Create temporary folder, only if it does not exist already
	root, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	sysFsRoot := filepath.Join(root, folder)
	if _, pathErr := os.Stat(sysFsRoot); pathErr != nil {
		err := os.Mkdir(sysFsRoot, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	config := Configuration{
		SysFsRoot: sysFsRoot,
		MQTT: MQTTConfig{
			Broker:   "mqtts://foo",
			ClientID: "unipitt",
			CAFile:   "foo",
		},
	}

	handler, err := NewHandler(config)
	if err == nil {
		t.Fatal("Expected an error with invalid caFile, got none")
	}
	defer handler.Close()
}

func TestHandlerCaFile(t *testing.T) {
	f, err := ioutil.TempFile("", "ca")
	if err != nil {
		t.Fatal(err)
	}
	caFile := f.Name()
	defer os.Remove(caFile)

	// Setup a folder structure
	folder := "di_1_01"
	// Create temporary folder, only if it does not exist already
	root, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	sysFsRoot := filepath.Join(root, folder)
	if _, pathErr := os.Stat(sysFsRoot); pathErr != nil {
		err := os.Mkdir(sysFsRoot, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}

	config := Configuration{
		SysFsRoot: sysFsRoot,
		MQTT: MQTTConfig{
			Broker:   "mqtts://foo",
			ClientID: "unipitt",
			CAFile:   caFile,
		},
	}
	handler, err := NewHandler(config)
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()
}
