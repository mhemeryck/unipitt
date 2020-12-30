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
	broker := "mqtts://foo"
	clientID := "unipitt"
	caFile := ""
	configFile := ""
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

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile)
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

func TestHandlerConfig(t *testing.T) {
	// Config file setup
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	content := []byte(`
topics:
  di_1_01: kitchen switch
  do_2_02: living light
`)
	if _, err := configFile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	expected := &Configuration{
		Topics: map[string]string{
			"di_1_01": "kitchen switch",
			"do_2_02": "living light",
		},
	}

	// Handler setup
	broker := "mqtts://foo"
	clientID := "unipitt"
	caFile := ""

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

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()

	// Check the topics mapped
	for k, v := range expected.Topics {
		if topic, ok := handler.config.Topics[k]; !ok {
			t.Errorf("Could not find topic for name %s\n", k)
		} else if topic != v {
			t.Errorf("Expected topic to be %s, but got %s\n", v, topic)
		}
	}
}

func TestHandlerConfigUnmarshalIssue(t *testing.T) {
	// Config file setup
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	content := []byte("foo")
	if _, err := configFile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	// Handler setup
	broker := "mqtts://foo"
	clientID := "unipitt"
	caFile := ""

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

	// Setup log monitoring
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()

	if !bytes.Contains(buf.Bytes(), []byte("Error reading config file")) {
		t.Fatal("Expected an error reading the config file, nothing happened")
	}
}

func TestHandlerNoSysFsRoot(t *testing.T) {
	broker := "mqtts://foo"
	clientID := "unipitt"
	caFile := ""
	configFile := ""

	sysFsRoot := "foo"

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile)
	if err == nil {
		t.Fatalf("Expected an error without a valid sys fs folder, got none")
	}
	defer handler.Close()
}

func TestHandlerCaFileIssue(t *testing.T) {
	broker := "mqtts://foo"
	clientID := "unipitt"
	caFile := "foo"
	configFile := ""

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

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile)
	if err == nil {
		t.Fatal("Expected an error with invalid caFile, got none")
	}
	defer handler.Close()
}

func TestHandlerCaFile(t *testing.T) {
	broker := "mqtts://foo"
	clientID := "unipitt"
	configFile := ""

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

	handler, err := NewHandler(broker, clientID, caFile, sysFsRoot, "", "", configFile)
	if err != nil {
		t.Fatal(err)
	}
	defer handler.Close()
}
