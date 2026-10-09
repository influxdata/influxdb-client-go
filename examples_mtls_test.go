// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2_test

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

func ExampleClient_configure_mtls() {
	serverCertPath := "path/to/server-certificate.crt"
	clientCertPath := "path/to/client-certificate.crt"
	clientKeyPath := "path/to/client-certificate.key"
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

	cert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		print(err)
		return
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
