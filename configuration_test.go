package unipitt

import (
	"testing"

	yaml "gopkg.in/yaml.v2"
)

func TestConfiguration(t *testing.T) {
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
