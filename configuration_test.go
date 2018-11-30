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

func TestConfigFromFileNonExistant(t *testing.T) {
	_, err := configFromFile("foo")
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
`)
	if _, err := configFile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := configFile.Close(); err != nil {
		t.Fatal(err)
	}

	c, err := configFromFile(configFile.Name())
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

	_, err = configFromFile(configFile.Name())
	if err == nil {
		t.Fatal("Expected an error on unmarshalling, got none")
	}
}
