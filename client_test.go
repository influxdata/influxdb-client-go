// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package influxdb2

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	ihttp "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/influxdata/influxdb-client-go/v2/domain"
	http2 "github.com/influxdata/influxdb-client-go/v2/internal/http"
	iwrite "github.com/influxdata/influxdb-client-go/v2/internal/write"
	ilog "github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"software.sslmate.com/src/go-pkcs12"
)

func TestUrls(t *testing.T) {
	urls := []struct {
		serverURL      string
		serverAPIURL   string
		writeURLPrefix string
	}{
		{"http://host:8086", "http://host:8086/api/v2/", "http://host:8086/api/v2/write"},
		{"http://host:8086/", "http://host:8086/api/v2/", "http://host:8086/api/v2/write"},
		{"http://host:8086/path", "http://host:8086/path/api/v2/", "http://host:8086/path/api/v2/write"},
		{"http://host:8086/path/", "http://host:8086/path/api/v2/", "http://host:8086/path/api/v2/write"},
		{"http://host:8086/path1/path2/path3", "http://host:8086/path1/path2/path3/api/v2/", "http://host:8086/path1/path2/path3/api/v2/write"},
		{"http://host:8086/path1/path2/path3/", "http://host:8086/path1/path2/path3/api/v2/", "http://host:8086/path1/path2/path3/api/v2/write"},
	}
	for _, url := range urls {
		t.Run(url.serverURL, func(t *testing.T) {
			c := NewClient(url.serverURL, "x")
			ci := c.(*clientImpl)
			assert.Equal(t, url.serverURL, ci.serverURL)
			assert.Equal(t, url.serverAPIURL, ci.httpService.ServerAPIURL())
			ws := iwrite.NewService("org", "bucket", ci.httpService, c.Options().WriteOptions())
			wu := ws.WriteURL()
			assert.Equal(t, url.writeURLPrefix+"?bucket=bucket&org=org&precision=ns", wu)
		})
	}
}

func TestWriteAPIManagement(t *testing.T) {
	data := []struct {
		org          string
		bucket       string
		expectedCout int
	}{
		{"o1", "b1", 1},
		{"o1", "b2", 2},
		{"o1", "b1", 2},
		{"o2", "b1", 3},
		{"o2", "b2", 4},
		{"o1", "b2", 4},
		{"o1", "b3", 5},
		{"o2", "b2", 5},
	}
	c := NewClient("http://localhost", "x").(*clientImpl)
	for i, d := range data {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			w := c.WriteAPI(d.org, d.bucket)
			assert.NotNil(t, w)
			assert.Len(t, c.writeAPIs, d.expectedCout)
			wb := c.WriteAPIBlocking(d.org, d.bucket)
			assert.NotNil(t, wb)
			assert.Len(t, c.syncWriteAPIs, d.expectedCout)
		})
	}
	c.Close()
	assert.Len(t, c.writeAPIs, 0)
	assert.Len(t, c.syncWriteAPIs, 0)
}

func TestUserAgentBase(t *testing.T) {
	ua := fmt.Sprintf("influxdb-client-go/%s (%s; %s)", Version, runtime.GOOS, runtime.GOARCH)
	assert.Equal(t, ua, http2.UserAgentBase)

}

type doer struct {
	userAgent string
	doer      ihttp.Doer
}

func (d *doer) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", d.userAgent)
	return d.doer.Do(req)
}

func TestUserAgent(t *testing.T) {
	ua := http2.UserAgentBase
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		if r.Header.Get("User-Agent") == ua {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	defer server.Close()
	var sb strings.Builder
	log.SetOutput(&sb)
	log.SetFlags(0)
	c := NewClientWithOptions(server.URL, "x", DefaultOptions().SetLogLevel(ilog.WarningLevel))
	assert.True(t, strings.Contains(sb.String(), "Application name is not set"))
	up, err := c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)

	c.Close()
	sb.Reset()
	// Test setting application  name
	c = NewClientWithOptions(server.URL, "x", DefaultOptions().SetApplicationName("Monitor/1.1"))
	ua = fmt.Sprintf("influxdb-client-go/%s (%s; %s) Monitor/1.1", Version, runtime.GOOS, runtime.GOARCH)
	assert.False(t, strings.Contains(sb.String(), "Application name is not set"))
	up, err = c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)
	c.Close()

	ua = "Monitor/1.1"
	opts := DefaultOptions()
	opts.HTTPOptions().SetHTTPDoer(&doer{
		userAgent: ua,
		doer:      http.DefaultClient,
	})

	//Create client with custom user agent setter
	c = NewClientWithOptions(server.URL, "x", opts)
	up, err = c.Ping(context.Background())
	require.NoError(t, err)
	assert.True(t, up)

	err = c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	assert.NoError(t, err)
	c.Close()
}

func TestServerError429(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		w.Header().Set("Retry-After", "1")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"code":"too many requests", "message":"exceeded rate limit"}`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "too many requests: exceeded rate limit", err.Error())
}

func TestServerOnPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proxy/0:0/influx/api/v2/write" {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(fmt.Sprintf(`{"code":"internal server error", "message":"%s"}`, r.URL.Path)))
		}
	}))

	defer server.Close()
	c := NewClient(server.URL+"/proxy/0:0/influx/", "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.NoError(t, err)
}

func TestServerErrorNonJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-time.After(100 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`internal server error`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	//Test non JSON error in custom code
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "500 Internal Server Error: internal server error", err.Error())

	// Test non JSON error from generated code
	params := &domain.GetBucketsParams{}
	b, err := c.APIClient().GetBuckets(context.Background(), params)
	assert.Nil(t, b)
	require.Error(t, err)
	assert.Equal(t, "500 Internal Server Error: internal server error", err.Error())

}

func TestServerErrorInflux1_8(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Influxdb-Error", "bruh moment")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "bruh moment"}`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "404 Not Found: bruh moment", err.Error())
}

func TestServerErrorEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	err := c.WriteAPIBlocking("o", "b").WriteRecord(context.Background(), "a,a=a a=1i")
	require.Error(t, err)
	assert.Equal(t, "Unexpected status code 404", err.Error())
}

func TestReadyFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html></html>`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	r, err := c.Ready(context.Background())
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestHealthFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`<html></html>`))
	}))

	defer server.Close()
	c := NewClient(server.URL, "x")
	h, err := c.Health(context.Background())
	assert.Error(t, err)
	assert.Nil(t, h)
}

const (
	influxdbServerCertFile = "internal/test/certificates/influxdb.crt"
	influxdbKeyFile        = "internal/test/certificates/influxdb.key"

	otherServerCertFile = "internal/test/certificates/other-server.crt"
	otherKeyFile        = "internal/test/certificates/other-server.key"
	otherCertP12        = "internal/test/certificates/other-server.p12"

	clientCertFile = "internal/test/certificates/client.crt"
	clientKeyFile  = "internal/test/certificates/client.key"
	clientCertP12  = "internal/test/certificates/client.p12"

	password = "changeit"
)

func TestTls(t *testing.T) {
	certificate, err := loadCertificate(influxdbServerCertFile, influxdbKeyFile, "")
	if err != nil {
		t.Fatalf("failed to load key pair: %v", err)
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "{}")
	}))
	defer ts.Close()
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	ts.StartTLS()

	testCases := []struct {
		name           string
		certPath       string
		expectError    bool
		expectedErrMsg string
	}{
		{
			name:        "valid server certificate",
			certPath:    influxdbServerCertFile,
			expectError: false,
		},
		{
			name:           "untrusted server certificate",
			certPath:       otherServerCertFile,
			expectError:    true,
			expectedErrMsg: "failed to verify certificate",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			certPool, err := loadX509CertPool(tc.certPath)
			if err != nil {
				assert.NoError(t, err)
			}

			tlsConfig := &tls.Config{
				RootCAs: certPool,
			}
			client := getClient(t, ts.URL, tlsConfig)
			_, err = client.Health(context.Background())
			if tc.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMutualTls(t *testing.T) {
	testCases := []struct {
		name           string
		certPath       string
		keyPath        string
		expectError    bool
		password       string
		expectedErrMsg string
	}{
		{
			name:        "valid client certificate",
			certPath:    clientCertFile,
			keyPath:     clientKeyFile,
			expectError: false,
		},
		{
			name:           "invalid client certificate send to the server",
			certPath:       otherServerCertFile,
			keyPath:        otherKeyFile,
			expectError:    true,
			expectedErrMsg: "tls: certificate required",
		},
		{
			name:        "valid client p12 certificate send to the server",
			certPath:    clientCertP12,
			password:    password,
			expectError: false,
		},
		{
			name:           "invalid client p12 certificate send to the server",
			certPath:       otherCertP12,
			password:       password,
			expectError:    true,
			expectedErrMsg: "tls: certificate required",
		},
	}

	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, "{}")
	}))
	defer ts.Close()

	cert, err := loadCertificate(influxdbServerCertFile, influxdbKeyFile, "")
	if err != nil {
		assert.NoError(t, err)
	}
	certPool, err := loadX509CertPool(clientCertFile)
	if err != nil {
		assert.NoError(t, err)
	}
	ts.TLS = &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	ts.StartTLS()

	for _, tc := range testCases {
		cert, err := loadCertificate(tc.certPath, tc.keyPath, password)
		if err != nil {
			assert.NoError(t, err)
		}

		t.Run(tc.name, func(t *testing.T) {
			certPool, err := loadX509CertPool(influxdbServerCertFile)
			if err != nil {
				assert.NoError(t, err)
			}
			tlsConfig := &tls.Config{
				RootCAs:      certPool,
				Certificates: []tls.Certificate{cert},
			}

			client := getClient(t, ts.URL, tlsConfig)
			_, err = client.Health(context.Background())
			if tc.expectError {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tc.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getClient(t *testing.T, url string, tlsConfig *tls.Config) Client {
	opts := DefaultOptions().SetTLSConfig(tlsConfig)
	client := NewClientWithOptions(url, "token", opts)
	t.Cleanup(func() {
		client.Close()
	})

	return client
}

func loadX509CertPool(certFile string) (*x509.CertPool, error) {
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(caCert)
	return pool, nil
}

func loadCertificate(certFile string, keyFile string, password string) (tls.Certificate, error) {
	if strings.HasSuffix(certFile, ".p12") {
		cert, err := processPKCS12(certFile, password)
		if err != nil {
			return tls.Certificate{}, err
		}
		return cert, nil
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return tls.Certificate{}, err
	}
	return cert, nil
}

func processPKCS12(path string, password string) (tls.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return tls.Certificate{}, err
	}
	privateKey, cert, caCerts, err := pkcs12.DecodeChain(data, password)
	if err != nil {
		return tls.Certificate{}, err
	}

	certBytes := [][]byte{cert.Raw}
	for _, ca := range caCerts {
		certBytes = append(certBytes, ca.Raw)
	}

	return tls.Certificate{
		Certificate: certBytes,
		PrivateKey:  privateKey,
		Leaf:        cert,
	}, nil
}
