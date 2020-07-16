package api

import (
	"context"
	"github.com/influxdata/influxdb-client-go/domain"
)

// QueryApi provides methods for performing synchronously flux query against InfluxDB server.
// Deprecated: Use QueryAPI instead.
type QueryApi interface {
	// QueryRaw executes flux query on the InfluxDB server and returns complete query result as a string with table annotations according to dialect
	QueryRaw(ctx context.Context, query string, dialect *domain.Dialect) (string, error)
	// Query executes flux query on the InfluxDB server and returns QueryTableResult which parses streamed response into structures representing flux table parts
	Query(ctx context.Context, query string) (*QueryTableResult, error)
}
