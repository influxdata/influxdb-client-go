package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	lp "github.com/influxdata/line-protocol"
)

func TestNewClient(t *testing.T) {
	type args struct {
		httpClient *http.Client
		options    []Option
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.httpClient, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_Ping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Header().Add("X-Influxdb-Version", "2.0-mock")
		if r.Method != "GET" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	//server.Start()
	defer server.Close()

	tests := []struct {
		name    string
		args    time.Duration
		want    time.Duration
		token   string
		org     string
		want1   string
		wantErr bool
	}{
		{
			name:  "ping",
			token: "faketoken",
			org:   "myorg",
			want1: "2.0-mock",
			args:  0,
		},
		{
			name:    "hanging ping",
			token:   "faketoken",
			org:     "myorg",
			args:    3 * time.Millisecond,
			want:    90 * time.Millisecond,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			cancel := func() {}
			if tt.args != 0 {
				ctx, cancel = context.WithTimeout(context.Background(), tt.args)
			} else {
				ctx = context.Background()
			}

			c, err := NewClient(server.Client(), WithToken(tt.token), WithAddress(server.URL))
			if err != nil {
				t.Error(err)
			}
			got, got1, err := c.Ping(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Client.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got > tt.want && tt.want != 0 {
				t.Errorf("Client.Ping() got = %v, want < %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Client.Ping() got1 = %v, want %v", got1, tt.want1)
			}
			cancel()
		})
	}
}

func TestClient_Write(t *testing.T) {
	type fields struct {
		httpClient           *http.Client
		contentEncoding      string
		gzipCompressionLevel int
		url                  *url.URL
		password             string
		username             string
		token                string
		org                  string
		maxRetries           int
		errOnFieldErr        bool
		userAgent            string
		authorization        string
	}
	type args struct {
		ctx    context.Context
		bucket string
		org    string
		m      []lp.Metric
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient:           tt.fields.httpClient,
				contentEncoding:      tt.fields.contentEncoding,
				gzipCompressionLevel: tt.fields.gzipCompressionLevel,
				url:                  tt.fields.url,
				password:             tt.fields.password,
				username:             tt.fields.username,
				token:                tt.fields.token,
				org:                  tt.fields.org,
				maxRetries:           tt.fields.maxRetries,
				errOnFieldErr:        tt.fields.errOnFieldErr,
				userAgent:            tt.fields.userAgent,
				authorization:        tt.fields.authorization,
			}
			if err := c.Write(tt.args.ctx, tt.args.bucket, tt.args.org, tt.args.m...); (err != nil) != tt.wantErr {
				t.Errorf("Client.Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
