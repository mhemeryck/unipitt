package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mhemeryck/unipitt"
)

// Current program version info, injected at build time
var version, commit, date string

const (
	// SysFsRoot default root folder to search for digital inputs
	SysFsRoot = "/sys/devices/platform/unipi_plc"
	// Payload default MQTT payload
	Payload = "trigger"
)

// printVersionInfo prints the current version info, where the values are injected at build time with goreleaser
func printVersionInfo() {
	log.Println("UnipiTT")
	var info = map[string]string{
		"Version": version,
		"Commit":  commit,
		"Date":    date,
	}
	for key, value := range info {
		log.Printf("%s: %s\n", key, value)
	}
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

func main() {
	// Arguments
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version info and exit")
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

	// Show version and exit
	if showVersion {
		printVersionInfo()
		return
	}
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
	readers, err := unipitt.FindDigitalInputReaders(sysfsRoot)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Created %d digital input reader instances from path %s\n", len(readers), sysfsRoot)
	for k := range readers {
		defer readers[k].Close()
	}

	events := make(chan *unipitt.DigitalInputReader)
	defer close(events)

	ticker := time.NewTicker(time.Duration(pollingInterval) * time.Millisecond)
	defer ticker.Stop()

	// Start polling
	for k := range readers {
		log.Printf("Initiate polling for %d readers\n", len(readers))
		go readers[k].Poll(events, ticker)
	}

	// Publish on a trigger
	for {
		select {
		case d := <-events:
			if d.Err != nil {
				log.Printf("Found error %s for topic %s\n", d.Err, d.Topic)
			} else {
				log.Printf("Trigger for topic %s\n", d.Topic)
				mqttClient.Publish(d.Topic, 0, false, payload)
			}
		}
	}
}
