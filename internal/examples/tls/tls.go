package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"software.sslmate.com/src/go-pkcs12"
)

func main() {
	// Config TLS with server a Certificate
	// Certification files can be one of this format .crt, .cert
	configTLS()

	// Config mTLS with client certificate send to the server for validating.
	// Using .crt and .key file format.
	configMutualTLSP12()

	// Config mTLS with both client certificate and server certificate.
	// Using .crt and .key file format.
	configMutualTLS()
}

func configTLS() {
	certFile := "path/to/server-certificate.crt"
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		print(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		RootCAs: certPool,
	}
	opts := influxdb2.DefaultOptions().SetTLSConfig(tlsConfig)
	client := influxdb2.NewClientWithOptions("https://localhost:8086", "token", opts)
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}

func configMutualTLSP12() {
	path := "path/to/client-certificate.p12"
	password := "password"

	data, err := os.ReadFile(path)
	if err != nil {

		print(err)
	}
	privateKey, cert, caCerts, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		print(err)
	}

	certBytes := [][]byte{cert.Raw}
	for _, ca := range caCerts {
		certBytes = append(certBytes, ca.Raw)
	}

	certificate := tls.Certificate{
		Certificate: certBytes,
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	opts := influxdb2.DefaultOptions().SetTLSConfig(tlsConfig)
	client := influxdb2.NewClientWithOptions("https://localhost:8086", "token", opts)
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}

func configMutualTLS() {
	serverCertPath := "path/to/server-certificate.crt"
	clientCertPath := "path/to/client-certificate.crt"
	clientKeyPath := "path/to/client-certificate.key"
	caCert, err := os.ReadFile(serverCertPath)
	if err != nil {
		print(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		print(err)
	}

	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
	}

	opts := influxdb2.DefaultOptions().SetTLSConfig(tlsConfig)
	client := influxdb2.NewClientWithOptions("https://localhost:8086", "token", opts)
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}
