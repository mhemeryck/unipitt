package unipitt

// Configuration represents the topic name for the MQTT message for a given instance name
type Configuration struct {
	Topics map[string]string
}
