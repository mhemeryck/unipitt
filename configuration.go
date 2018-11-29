package unipitt

import (
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Configuration represents the topic name for the MQTT message for a given instance name
type Configuration struct {
	Topics map[string]string
}

// Topic gets a topic (value) for a given name (key). Return the name itself as fallback
func (c *Configuration) Topic(name string) string {
	if value, ok := c.Topics[name]; ok {
		return value
	}
	return name
}

// reverseTopics construct reverse mapping of topics
func (c *Configuration) reverseTopics() map[string]string {
	r := make(map[string]string)
	for key, value := range c.Topics {
		r[value] = key
	}
	return r
}

// Name reverse mapping of topic for given name. In case nothing is found, just return the topic itself, hoping there's a mapped instance for it
func (c *Configuration) Name(topic string) string {
	if name, ok := c.reverseTopics()[topic]; ok {
		return name
	}
	return topic
}

// configFromFile reads a configuration from a yaml file
func configFromFile(configFile string) (c Configuration, err error) {
	f, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("Error reading config file: %s\n", err)
		return
	}

	err = yaml.Unmarshal(f, &c)
	if err != nil {
		log.Printf("Error unmarshalling the config: %s\n", err)
		return
	}
	return
}
