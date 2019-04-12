package influxdb

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	cmp "github.com/google/go-cmp/cmp"
)

func TestNew(t *testing.T) {
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
		{
			name: "basic",
			args: args{
				httpClient: http.DefaultClient,
				options: []Option{
					WithToken("foo"),
				},
			},
			want: &Client{
				httpClient:       http.DefaultClient,
				authorization:    "Token foo",
				contentEncoding:  "gzip",
				compressionLevel: 4,
				url: func() *url.URL {
					u, err := url.Parse("http://127.0.0.1:9999/api/v2")
					if err != nil {
						t.Fatal(err)
					}
					return u
				}(),
				userAgent: ua(),
			},
		},
		{
			name: "no compression",
			args: args{
				httpClient: http.DefaultClient,
				options: []Option{
					WithToken("foo"),
					WithNoCompression(),
				},
			},
			want: &Client{
				httpClient:       http.DefaultClient,
				authorization:    "Token foo",
				contentEncoding:  "",
				compressionLevel: 4,
				url: func() *url.URL {
					u, err := url.Parse("http://127.0.0.1:9999/api/v2")
					if err != nil {
						t.Fatal(err)
					}
					return u
				}(),
				userAgent: ua(),
			},
		},
		{
			name: "custom ua",
			args: args{
				httpClient: http.DefaultClient,
				options: []Option{
					WithToken("foo"),
					WithUserAgent("fake-user-agent"),
				},
			},
			want: &Client{
				httpClient:       http.DefaultClient,
				authorization:    "Token foo",
				contentEncoding:  "gzip",
				compressionLevel: 4,
				url: func() *url.URL {
					u, err := url.Parse("http://127.0.0.1:9999/api/v2")
					if err != nil {
						t.Fatal(err)
					}
					return u
				}(),
				userAgent: "fake-user-agent",
			},
		},
		{
			name: "compression level",
			args: args{
				httpClient: http.DefaultClient,
				options: []Option{
					WithToken("foo"),
					WithUserAgent("fake-user-agent"),
					WithGZIP(6),
				},
			},
			want: &Client{
				httpClient:       http.DefaultClient,
				authorization:    "Token foo",
				contentEncoding:  "gzip",
				compressionLevel: 6,
				url: func() *url.URL {
					u, err := url.Parse("http://127.0.0.1:9999/api/v2")
					if err != nil {
						t.Fatal(err)
					}
					return u
				}(),
				userAgent: "fake-user-agent",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.httpClient, tt.args.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmp.AllowUnexported(Client{})) {
				t.Errorf("Diff: %s", cmp.Diff(got, tt.want, cmp.AllowUnexported(Client{})))
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

			c, err := New(server.Client(), WithToken(tt.token), WithAddress(server.URL))
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
