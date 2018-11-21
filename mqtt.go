package unipitt

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
)

// NewTLSConfig generates a TLS config instance for use with the MQTT setup
func NewTLSConfig(caFile string) (*tls.Config, error) {
	// Read the ceritifcates from the system, continue with empty pool in case of failure
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read the local file from the supplied path
	certs, err := ioutil.ReadFile(caFile)
	if err != nil {
		log.Printf("Failed to append %q to RootCAs: %v", caFile, err)
	}
	// Append our cert to the system pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		log.Println("No certs appended, using system certs only")
	}

	// Trust the augmented cert pool in our client
	return &tls.Config{RootCAs: rootCAs}, err
}
