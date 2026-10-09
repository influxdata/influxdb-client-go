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
	// Configure TLS with a custom server CA certificate (.crt, .cert).
	configureTLSWithServerCert()

	// Configure mutual TLS (mTLS) for client authentication using a PKCS#12 (.p12) bundle.
	configureMutualTLSWithPKCS12()

	// Configure mutual TLS (mTLS) using PEM certificate and private key files (.crt, .key).
	// This example also includes a server certificate so the client can validate the server identity. This is optional.
	configureMutualTLSWithPEM()
}

func configureTLSWithServerCert() {
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
	defer client.Close()
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}

func configureMutualTLSWithPKCS12() {
	path := "path/to/client-certificate.p12"
	password := "password"

	data, err := os.ReadFile(path)
	if err != nil {

		print(err)
	}
	privateKey, cert, caCerts, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		print(err)
		return
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
	defer client.Close()
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}

func configureMutualTLSWithPEM() {
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
	defer client.Close()
	_, err = client.Health(context.Background())
	if err != nil {
		print(err)
	}
}
