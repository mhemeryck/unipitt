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
	config    Configuration
}

// NewHandler prepares and sets up an entire unipitt handler
func NewHandler(config Configuration) (h *Handler, err error) {
	h = &Handler{}

        h.config = config

	// Digital writer setup
	h.writerMap, err = FindDigitalOutputWriters(h.config.SysFsRoot)
	if err != nil {
		log.Printf("Error creating a map of digital output writers: %s\n", err)
	}

	log.Printf("connecting to mqtt server with options: %s\n", h.config.Mqtt)
	// MQTT setup
	opts := mqtt.NewClientOptions()
	opts.AddBroker(h.config.Mqtt.Broker)
	opts.SetClientID(h.config.Mqtt.ClientID)
	opts.SetUsername(h.config.Mqtt.Username)
	opts.SetPassword(h.config.Mqtt.Password)
	if h.config.Mqtt.CAFile != "" {
		tlsConfig, err := NewTLSConfig(h.config.Mqtt.CAFile)
		if err != nil {
			log.Printf("Error reading MQTT CA file %s: %s\n", h.config.Mqtt.CAFile, err)
			return h, err
		}
		opts.SetTLSConfig(tlsConfig)
	}

	// Callbacks for subscribe
	var cb mqtt.MessageHandler = func(c mqtt.Client, msg mqtt.Message) {
		log.Printf("Handling message on topic %s\n", msg.Topic())
		// Find corresponding writer
		if writer, ok := h.writerMap[h.config.Name(msg.Topic())]; ok {
			err := writer.Update(string(msg.Payload()) == MsgTrueValue)
			if err != nil {
				log.Printf("Error updating digital output with name %s: %s\n", writer.Name, err)
			}
		} else {
			log.Printf("Error matching a writer for given name %s\n", msg.Topic())
		}
	}
	opts.OnConnect = func(c mqtt.Client) {
		for name := range h.writerMap {
			// Also subscribe any given mapped topic for the names
			topic := h.config.Topic(name)
			if token := c.Subscribe(topic, 0, cb); token.Wait() && token.Error() != nil {
				log.Print(token.Error())
			}
		}
	}

	h.client = mqtt.NewClient(opts)
	err = h.connect()
	if err != nil {
		log.Printf("Error connecting to MQTT broker: %s\n ...", err)
	}

	// Digital Input reader setup
	h.readers, err = FindDigitalInputReaders(h.config.SysFsRoot)
	if err != nil {
		return
	}
	log.Printf("Created %d digital input reader instances from path %s\n", len(h.readers), h.config.SysFsRoot)

	return
}

// Poll starts the actual polling and pushing to MQTT
func (h *Handler) Poll(done chan bool, interval int) (err error) {
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
				log.Printf("Found error %s for name %s\n", d.Err, d.Name)
			} else {
			        var payload string
				// Determine topic from config
				log.Printf("Trigger for name %s, using topic %s\n", d.Name, h.config.Topic(d.Name))
				if (d.Value) {
					payload = "ON"
				} else {
					payload = "OFF"
				}
				if token := h.client.Publish(h.config.Topic(d.Name), 0, false, payload); token.Wait() && token.Error() != nil {
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
