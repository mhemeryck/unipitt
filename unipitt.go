package unipitt

import (
	"log"

	"github.com/cenkalti/backoff"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// SysFsRoot default root folder to search for digital inputs
	SysFsRoot = "/sys/devices/platform/unipi_plc"
)

// Unipitt defines the interface with unipi board
type Unipitt interface {
	Poll(pollingInterval int) error
	Close()
}

// Handler implements handles all unipi to MQTT interactions
type Handler struct {
	readers []DigitalInputReader
	client  mqtt.Client
}

// NewHandler prepares and sets up an entire unipitt handler
func NewHandler(broker string, clientID string, caFile string, sysFsRoot string) (h *Handler, err error) {
	h = &Handler{}
	// MQTT setup
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	if caFile != "" {
		tlsConfig, err := NewTLSConfig(caFile)
		if err != nil {
			log.Printf("Error reading MQTT CA file %s: %s\n", caFile, err)
		} else {
			opts.SetTLSConfig(tlsConfig)
		}
	}
	h.client = mqtt.NewClient(opts)
	err = h.connect()
	if err != nil {
		log.Println("Error connecting to MQTT broker ...")
	}

	// Digital Input reader setup
	h.readers, err = FindDigitalInputReaders(sysFsRoot)
	if err != nil {
		return
	}
	log.Printf("Created %d digital input reader instances from path %s\n", len(h.readers), sysFsRoot)
	return
}

// Poll starts the actual polling and pushing to MQTT
func (h *Handler) Poll(done chan bool, interval int, payload string) (err error) {
	events := make(chan *DigitalInputReader)
	defer close(events)

	// Start polling
	log.Printf("Initiate polling for %d readers\n", len(h.readers))
	for k := range h.readers {
		go h.readers[k].Poll(events, interval)
	}

	// Publish on a trigger
	for {
		select {
		case d := <-events:
			if d.Err != nil {
				log.Printf("Found error %s for topic %s\n", d.Err, d.Topic)
			} else {
				log.Printf("Trigger for topic %s\n", d.Topic)
				if token := h.client.Publish(d.Topic, 0, false, payload); token.Wait() && token.Error() != nil {
					go backoff.Retry(h.connect, backoff.NewExponentialBackOff())
				}
			}
		case <-done:
			log.Println("Handler done polling, coming back ...")
			return
		}
	}
}

// reconnect tries to reconnect the MQTT client to the broker
func (h *Handler) connect() error {
	log.Println("Error connecting to MQTT broker ...")
	token := h.client.Connect()
	token.Wait()
	return token.Error()
}

// Close loose ends
func (h *Handler) Close() {
	// Close the readers
	for k := range h.readers {
		h.readers[k].Close()
	}
}
