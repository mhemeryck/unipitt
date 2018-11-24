package unipitt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewDigitalOutputWriter(t *testing.T) {
	folder := "foo/bar"
	d := NewDigitalOutputWriter(folder)
	if d.Topic != "bar" {
		t.Fatalf("Expected topic %s, got %s\n", "bar", d.Topic)
	}
	if d.Path != folder {
		t.Fatalf("Expected path %s, got %s\n", folder, d.Path)
	}
}

func TestUpdateDigitalOutputWriter(t *testing.T) {
	folder := "do_2_01"

	// Create temporary folder, only if it does not exist already
	sysFsRoot, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sysFsRoot)
	doFolder := filepath.Join(sysFsRoot, folder)
	if _, pathErr := os.Stat(doFolder); pathErr != nil {
		err := os.Mkdir(doFolder, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}
	fname := filepath.Join(doFolder, DoFilename)

	d := NewDigitalOutputWriter(doFolder)

	cases := []struct {
		Given    bool
		Expected string
	}{
		{Given: false, Expected: DoFalseValue},
		{Given: true, Expected: DoTrueValue},
	}
	for _, testCase := range cases {
		err := d.Update(testCase.Given)
		if err != nil {
			t.Fatal(err)
		}
		f, err := os.Open(fname)
		if err != nil {
			t.Fatal(err)
		}
		b := make([]byte, 2)
		f.Read(b)
		if !bytes.Equal(b, []byte(testCase.Expected)) {
			t.Fatalf("Expected %s, got %s\n", testCase.Expected, string(b))
		}
	}
}

func TestUpdateDigitalOutputWriterBogusFolder(t *testing.T) {
	folder := "/foo/bar"
	d := NewDigitalOutputWriter(folder)
	err := d.Update(false)
	if err == nil {
		t.Fatal("Expected an error, found none")
	}
}

func TestFindDigitalOutputWriters(t *testing.T) {
	folder := "do_2_01"

	// Create temporary folder, only if it does not exist already
	sysFsRoot, err := ioutil.TempDir("", "unipitt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(sysFsRoot)
	doFolder := filepath.Join(sysFsRoot, folder)
	if _, pathErr := os.Stat(doFolder); pathErr != nil {
		err := os.Mkdir(doFolder, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}
	}
	writerMap, err := FindDigitalOutputWriters(sysFsRoot)
	if err != nil {
		t.Fatal(err)
	}
	if writer, ok := writerMap[folder]; !ok {
		t.Fatalf("Expected to find writer with topic %s in map for name %s\n", writer.Topic, folder)
	}

}

func TestFindDigitalOutputWritersNoFolder(t *testing.T) {
	folder := "foo"

	_, err := FindDigitalOutputWriters(folder)
	if err == nil {
		t.Fatal("Expected no mapping to be found")
	}
}
