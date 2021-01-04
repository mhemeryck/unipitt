package unipitt

import (
	"io/ioutil"
	"os"
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestConfigurationUnmarshal(t *testing.T) {
	input := []byte(`
topics:
  di_1_01: kitchen switch
  do_2_02: living light
`)
	var c Configuration
	err := yaml.Unmarshal(input, &c)
	if err != nil {
		t.Fatal(err)
	}

	expected := &Configuration{
		Topics: map[string]string{
			"di_1_01": "kitchen switch",
			"do_2_02": "living light",
		},
	}

	for k, v := range expected.Topics {
		if topic, ok := c.Topics[k]; !ok {
			t.Errorf("Could not find topic for name %s\n", k)
		} else if topic != v {
			t.Errorf("Expected topic to be %s, but got %s\n", v, topic)
		}
	}
}

func TestConfigurationTopic(t *testing.T) {
	cases := []struct {
		Name     string
		Expected string
	}{
		{Name: "foo", Expected: "qux"},
		{Name: "bar", Expected: "bar"},
	}
	c := Configuration{
		Topics: map[string]string{
			"foo": "qux",
		},
	}
	for _, testCase := range cases {
		result := c.Topic(testCase.Name)
		if result != testCase.Expected {
			t.Fatalf("Expected result to be %s, got %s\n", testCase.Expected, result)
		}
	}
}

func TestConfigurationTopicWithPrefix(t *testing.T) {
	cases := []struct {
		Name     string
		Expected string
	}{
		{Name: "foo", Expected: "/home/qux"},
		{Name: "bar", Expected: "/home/bar"},
	}
	c := Configuration{
		Topics: map[string]string{
			"foo": "qux",
		},
		Mqtt: MqttConfig{
			TopicPrefix: "/home/",
		},
	}
	for _, testCase := range cases {
		result := c.Topic(testCase.Name)
		if result != testCase.Expected {
			t.Fatalf("Expected result to be %s, got %s\n", testCase.Expected, result)
		}
	}
}

func TestConfigurationName(t *testing.T) {
	cases := []struct {
		Topic    string
		Expected string
	}{
		{Topic: "foo", Expected: "qux"},
		{Topic: "bar", Expected: "bar"},
	}
	c := Configuration{
		Topics: map[string]string{
			"qux": "foo",
		},
	}
	for _, testCase := range cases {
		result := c.Name(testCase.Topic)
		if result != testCase.Expected {
			t.Fatalf("Expected result to be %s, got %s\n", testCase.Expected, result)
		}
	}
}

func TestConfigurationNameWithPrefix(t *testing.T) {
	cases := []struct {
		Topic    string
		Expected string
	}{
		{Topic: "/home/foo", Expected: "qux"},
		{Topic: "/home/bar", Expected: "bar"},
		
		// this should not happen as the topic is not starting with
		// the configured prefix
		{Topic: "huh", Expected: "huh"},
	}
	c := Configuration{
		Topics: map[string]string{
			"qux": "foo",
		},
		Mqtt: MqttConfig{
			TopicPrefix: "/home/",
		},
	}
	for _, testCase := range cases {
		result := c.Name(testCase.Topic)
		if result != testCase.Expected {
			t.Fatalf("Expected result to be %s, got %s\n", testCase.Expected, result)
		}
	}
}

func TestConfigFromFileNonExistant(t *testing.T) {
	_, err := ReadConfigFromFile("foo")
	if err == nil {
		t.Fatalf("Expected an error to occur reading non-existent file, got not none")
	}
}

func TestConfigFromFile(t *testing.T) {
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	content := []byte(`
topics:
  di_1_01: kitchen switch
  do_2_02: living light
mqtt:
  username: mqttuser
  password: mqttpass
  client_id: foobar
  topic_prefix: /prefix/
  ca_file: pathtocafile
`)
	if _, err := configFile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	c, err := ReadConfigFromFile(configFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	expected := &Configuration{
		Topics: map[string]string{
			"di_1_01": "kitchen switch",
			"do_2_02": "living light",
		},
		Mqtt: MqttConfig{
			Username: "mqttuser",
			Password: "mqttpass",
			ClientID: "foobar",
			TopicPrefix: "/prefix/",
			CAFile: "pathtocafile",
		},
	}

	for k, v := range expected.Topics {
		if topic, ok := c.Topics[k]; !ok {
			t.Errorf("Could not find topic for name %s\n", k)
		} else if topic != v {
			t.Errorf("Expected topic to be %s, but got %s\n", v, topic)
		}
	}
	if c.Mqtt != expected.Mqtt {
		t.Errorf("Expected Mqtt parameters to be %s but got %s\n", expected.Mqtt, c.Mqtt)
	}

}

func TestConfigFromFileUnmarshalIssue(t *testing.T) {
	configFile, err := ioutil.TempFile("", "config")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(configFile.Name())

	content := []byte("foo")

	if _, err := configFile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = ReadConfigFromFile(configFile.Name())
	if err == nil {
		t.Fatal("Expected an error on unmarshalling, got none")
	}
}
