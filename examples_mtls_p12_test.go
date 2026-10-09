// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "crypto/x509"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"software.sslmate.com/src/go-pkcs12"
)

func ExampleClient_configure_mtls_with_p12_file() {
	serverCertPath := "path/to/server-certificate.crt"
	path := "path/to/client-certificate.p12"
	password := "password"

	data, err := os.ReadFile(path)
	if err != nil {
		print(err)
		return
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

	caCert, err := os.ReadFile(serverCertPath)
	if err != nil {
		print(err)
		return
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		print("failed to parse server certificate")
		return
	}

	certificate := tls.Certificate{
		Certificate: certBytes,
		PrivateKey:  privateKey,
		Leaf:        cert,
	}

	tlsConfig := &tls.Config{
		RootCAs:      certPool,
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
