package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// Filename to check for
	Filename = "di_value"
	// TrueValue is the value considered to be true
	TrueValue = "1"
	// FolderRegex represents to regular expression used for finding the required file to read from
	FolderRegex = "di_[0-9]_[0-9]{2}"
	// SysFsRoot default root folder to search for digital inputs
	SysFsRoot = "/sys/devices/platform/unipi_plc"
	// Payload default MQTT payload
	Payload = "trigger"
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
func (d *DigitalInputReader) Poll(events chan *DigitalInputReader, pollingInterval int) (err error) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(pollingInterval)).C
	for {
		select {
		case <-ticker:
			err = d.Update(events)
			if err != nil {
				return
			}
		}
	}
}

// Close closes the current open file handle
func (d *DigitalInputReader) Close() {
	d.f.Close()
}

// NewTLSConfig generates a TLS config instance for use with the MQTT setup
func NewTLSConfig(caFile string) *tls.Config {
	// Read the ceritifcates from the system, continue with empty pool in case of failure
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read the local file from the supplied path
	certs, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Fatalf("Failed to append %q to RootCAs: %v", caFile, err)
	}
	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	return &tls.Config{
		RootCAs: rootCAs,
	}
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

func main() {
	// Arguments
	var pollingInterval int
	flag.IntVar(&pollingInterval, "polling_interval", 50, "Polling interval for one coil group in millis")
	var caFile string
	flag.StringVar(&caFile, "cafile", "", "CA certificate used for MQTT TLS setup")
	var broker string
	flag.StringVar(&broker, "broker", "ssl://raspberrypi.lan:8883", "MQTT broker URI")
	var clientID string
	flag.StringVar(&clientID, "client_id", "unipitt", "MQTT host client ID")
	var sysfsRoot string
	flag.StringVar(&sysfsRoot, "sysfs_root", SysFsRoot, "Root folder to search for digital inputs")
	var payload string
	flag.StringVar(&payload, "payload", Payload, "Default MQTT message payload")
	flag.Parse()

	// MQTT setup
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	if caFile != "" {
		tlsConfig := NewTLSConfig(caFile)
		opts.SetTLSConfig(tlsConfig)
	}
	mqttClient := mqtt.NewClient(opts)
	token := mqttClient.Connect()
	if token.Wait() && token.Error() != nil {
		log.Println("Can't connect to MQTT host")
	}

	// Setup digital input
	paths, err := findDigitalInputPaths(sysfsRoot)
	if err != nil {
		log.Fatalln(err)
	}
	readers := make([]DigitalInputReader, len(paths))
	for k, folder := range paths {
		// Read topic as the trailing folder path
		_, topic := path.Split(folder)
		digitalInputReader, err := NewDigitalInputReader(folder, topic)
		if err != nil {
			log.Print(err)
		}
		defer digitalInputReader.Close()
		readers[k] = *digitalInputReader
	}

	events := make(chan *DigitalInputReader)

	// Start polling
	for k := range readers {
		go readers[k].Poll(events, pollingInterval)
	}

	// Publish on a trigger
	for {
		select {
		case d := <-events:
			log.Printf("Trigger for topic %s\n", d.Topic)
			mqttClient.Publish(d.Topic, 0, false, payload)
		}
	}
}
