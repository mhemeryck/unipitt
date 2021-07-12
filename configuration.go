package unipitt

import (
	"io/ioutil"
	"log"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// Configuration represents the topic name for the MQTT message for a given instance name

type MQTTConfig struct {
	Broker      string
	CAFile      string `yaml:"ca_file"`
	ClientID    string `yaml:"client_id"`
	Username    string
	Password    string
	TopicPrefix string `yaml:"topic_prefix"`
}

type Configuration struct {
	Topics    map[string]string
	MQTT      MQTTConfig
	SysFsRoot string `yaml:"sys_fs_root"`
}

// Topic gets a topic (value) for a given name (key). Return the name itself as fallback
func (c *Configuration) Topic(name string, suffix string) string {
	if value, ok := c.Topics[name]; ok {
		return c.MQTT.TopicPrefix + value + suffix
	}
	return c.MQTT.TopicPrefix + name + suffix
}

// reverseTopics construct reverse mapping of topics
func (c *Configuration) reverseTopics(suffix string) map[string]string {
	r := make(map[string]string)
	for key, value := range c.Topics {
		r[c.MQTT.TopicPrefix+value+suffix] = key
	}
	return r
}

// Name reverse mapping of topic for given name. In case nothing is found, just return the topic itself, hoping there's a mapped instance for it
func (c *Configuration) Name(topic string, suffix string) string {
	if name, ok := c.reverseTopics(suffix)[topic]; ok {
		return name
	}
	
	if strings.HasSuffix(topic, suffix) {
		topic = topic[:len(topic)-len(suffix)]
	}
	if strings.HasPrefix(topic, c.MQTT.TopicPrefix) {
		topic = topic[len(c.MQTT.TopicPrefix):]
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
