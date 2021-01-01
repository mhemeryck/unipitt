package unipitt

import (
	"io/ioutil"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Configuration represents the topic name for the MQTT message for a given instance name

type MqttConfig struct {
	Broker string
	CAFile string
	ClientID string
	Username string
	Password string
	TopicPrefix string
}

type Configuration struct {
	Topics map[string]string
	Mqtt MqttConfig
	SysFsRoot string
}

// Topic gets a topic (value) for a given name (key). Return the name itself as fallback
func (c *Configuration) Topic(name string) string {
	if value, ok := c.Topics[name]; ok {
		return c.Mqtt.TopicPrefix + value
	}
	return c.Mqtt.TopicPrefix + name
}

// reverseTopics construct reverse mapping of topics
func (c *Configuration) reverseTopics() map[string]string {
	r := make(map[string]string)
	for key, value := range c.Topics {
		r[c.Mqtt.TopicPrefix + value] = key
	}
	return r
}

// Name reverse mapping of topic for given name. In case nothing is found, just return the topic itself, hoping there's a mapped instance for it
func (c *Configuration) Name(topic string) string {
	if name, ok := c.reverseTopics()[topic]; ok {
		return name
	}
	if strings.HasPrefix(topic, c.Mqtt.TopicPrefix) {
		return topic[len(c.Mqtt.TopicPrefix):]
	}
	return topic
}

// UpdateConfigFromFile updates the configuration with values from a yaml file
func UpdateConfigFromFile(configFile string, c *Configuration) (err error) {
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

// ReadConfigFromFile reads a configuration from a yaml file
func ReadConfigFromFile(configFile string) (c Configuration, err error) {

	err = UpdateConfigFromFile(configFile, &c)
	if err != nil {
		return
	}
	return
}
