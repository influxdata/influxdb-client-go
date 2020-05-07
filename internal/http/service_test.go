package http_test

import (
	ihttp "github.com/influxdata/influxdb-client-go/internal/http"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServiceImpl_ServerUrl(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		serverUrl string
	}{
		{
			name:      "without trailing slash",
			url:       "http://localhost:9999",
			serverUrl: "http://localhost:9999",
		},
		{
			name:      "with trailing slash",
			url:       "http://localhost:9999/",
			serverUrl: "http://localhost:9999",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := ihttp.NewService(tt.url, "Token my-token", nil, 20)
			require.Equal(t, tt.serverUrl, service.ServerUrl())
		})
	}
}

func TestServiceImpl_ServerApiUrl(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		serverApiUrl string
	}{
		{
			name:         "without trailing slash",
			url:          "http://localhost:9999",
			serverApiUrl: "http://localhost:9999/api/v2/",
		},
		{
			name:         "with trailing slash",
			url:          "http://localhost:9999/",
			serverApiUrl: "http://localhost:9999/api/v2/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := ihttp.NewService(tt.url, "Token my-token", nil, 20)
			require.Equal(t, tt.serverApiUrl, service.ServerApiUrl())
		})
	}
}
