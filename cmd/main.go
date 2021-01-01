package main

import (
	"flag"
	"log"

	"github.com/mhemeryck/unipitt"
)

// Current program version info, injected at build time
var version, commit, date string

const (
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

func main() {
	// Arguments
	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version info and exit")
	var pollingInterval int
	flag.IntVar(&pollingInterval, "polling_interval", 50, "Polling interval per digital input in millis")
	var caFile string
	flag.StringVar(&caFile, "cafile", "", "CA certificate used for MQTT TLS setup")
	var broker string
	flag.StringVar(&broker, "broker", "ssl://raspberrypi.lan:8883", "MQTT broker URI")
	var clientID string
	flag.StringVar(&clientID, "client_id", "unipitt", "MQTT host client ID")
	var username string
	flag.StringVar(&username, "username", "", "MQTT username")
	var password string
	flag.StringVar(&password, "password", "", "MQTT password")	
	var sysFsRoot string
	flag.StringVar(&sysFsRoot, "sysfs_root", unipitt.SysFsRoot, "Root folder to search for digital inputs")
	var configFile string
	flag.StringVar(&configFile, "config", "", "Config file name")
	flag.Parse()

	// Show version and exit
	if showVersion {
		printVersionInfo()
		return
	}

	// Initialize config from command line
	config := unipitt.Configuration{
		SysFsRoot: sysFsRoot,
		Mqtt: unipitt.MqttConfig{
			Broker: broker,
			ClientID: clientID,
			CAFile: caFile,
			Username: username,
			Password: password,
		},
	}

	// Check if there's a config to be read
	if configFile != "" {
		log.Printf("Reading configuration file %s\n", configFile)
		err := unipitt.UpdateConfigFromFile(configFile, &config)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Setup handler
	handler, err := unipitt.NewHandler(config)
	if err != nil {
		log.Fatal(err)
	}
	defer handler.Close()

	// Start polling (blocking)
	done := make(chan bool)
	defer close(done)
	handler.Poll(done, pollingInterval)
}
