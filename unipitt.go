package unipitt

import (
	"log"

	"github.com/cenkalti/backoff"
	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// SysFsRoot default root folder to search for digital inputs
	SysFsRoot = "/sys/devices/platform/unipi_plc"
	// MsgTrueValue is the MQTT true value to check for
	MsgTrueValue = "ON"
)

// Unipitt defines the interface with unipi board
type Unipitt interface {
	Poll(pollingInterval int) error
	Close()
}

// Handler implements handles all unipi to MQTT interactions
type Handler struct {
	readers   []DigitalInputReader
	writerMap map[string]DigitalOutputWriter
	client    mqtt.Client
}

// NewHandler prepares and sets up an entire unipitt handler
func NewHandler(broker string, clientID string, caFile string, sysFsRoot string) (h *Handler, err error) {
	h = &Handler{}
	// Digital writer setup
	// Set message handler as callback
	h.writerMap, err = FindDigitalOutputWriters(sysFsRoot)
	if err != nil {
		log.Printf("Error creating a map of digital output writers: %s\n", err)
	}

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

	// Callbacks for subscribe
	var cb mqtt.MessageHandler = func(c mqtt.Client, msg mqtt.Message) {
		if writer, ok := h.writerMap[msg.Topic()]; ok {
			err := writer.Update(string(msg.Payload()) == MsgTrueValue)
			if err != nil {
				log.Printf("Error updating digital output with topic %s: %s\n", writer.Topic, err)
			}
		} else {
			log.Printf("Error matching a writer for given topic %s\n", msg.Topic())
		}
	}
	opts.OnConnect = func(c mqtt.Client) {
		for topic := range h.writerMap {
			if token := c.Subscribe(topic, 0, cb); token.Wait() && token.Error() != nil {
				log.Print(err)
			}
		}
	}

	h.client = mqtt.NewClient(opts)
	err = h.connect()
	if err != nil {
		log.Printf("Error connecting to MQTT broker: %s\n ...", err)
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
