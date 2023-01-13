// Copyright 2020-2023 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/apache/arrow/go/v10/arrow/flight"
	"github.com/apache/arrow/go/v10/arrow/flight/flightsql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type grpcCredentials struct {
	token      string
	bucketName string
}

func (g grpcCredentials) GetRequestMetadata(ctx context.Context, in ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + g.token,
		"bucket-name":   g.bucketName,
	}, nil
}

func (grpcCredentials) RequireTransportSecurity() bool {
	return true
}

type QuerySQLAPI interface {
	Query(ctx context.Context, bucket string, query string) (*flight.Reader, error)
}

// NewQuerySQLAPI returns new query client for querying with SQL
func NewQuerySQLAPI(hostAndPort string, token string) QuerySQLAPI {
	return &querySQLAPI{
		hostAndPort: hostAndPort,
		token:       token,
	}
}

// queryAPI implements QueryAPI interface
type querySQLAPI struct {
	hostAndPort string
	token       string
}

// Query returns an slice of Arrow records from an SQL query to a bucket
func (q *querySQLAPI) Query(ctx context.Context, bucket string, query string) (*flight.Reader, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, fmt.Errorf("error loading system certificates: %v", err)
	}

	flightClient, err := flightsql.NewClient(
		q.hostAndPort, nil, nil,
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		grpc.WithPerRPCCredentials(grpcCredentials{token: q.token, bucketName: bucket}),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("error creating flightsql client: %v", err)
	}

	flightInfo, err := flightClient.Execute(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %v", err)
	}

	reader, err := flightClient.DoGet(ctx, flightInfo.Endpoint[0].Ticket)
	if err != nil {
		return nil, fmt.Errorf("error getting flight tickets: %v", err)
	}

	return reader, nil
}
