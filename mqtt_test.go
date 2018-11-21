package unipitt

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestNewTLSConfig(t *testing.T) {
	f, err := ioutil.TempFile("", "ca")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	c, err := NewTLSConfig(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if c.RootCAs == nil {
		t.Fatal("Expected config to have some root CA, got none")
	}
}

func TestNewTLSConfigNoFile(t *testing.T) {
	_, err := NewTLSConfig("foo")
	if err == nil {
		t.Fatal("Expected an error reading the file, got none")
	}
}
