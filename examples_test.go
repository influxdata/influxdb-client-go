package influxdb_test

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"text/template"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
)

func writeHandler(w http.ResponseWriter, r *http.Request) {
	reader := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			log.Fatal(err)
		}
	}
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
	w.WriteHeader(200)
}

func setupHandler(w http.ResponseWriter, r *http.Request) {
	// this is an easy way of making the result
	// it works, but doesn't exscape properly
	// its only this way because...  mock
	resTemplate, err := template.New("result").Parse(`{"user":{"links":{"logs":"/api/v2/users/03b00a760bde3000/logs","self":"/api/v2/users/03b00a760bde3000"},"id":"03b00a760bde3000","name":"{{ .username }}"},"bucket":{"id":"03b00a761e9e3001","organizationID":"03b00a761e9e3000","organization":"{{ .org }}","name":"{{ .org }}-bucket","retentionRules":[{"type":"expire","everySeconds":{{ .retentionSeconds }}}],"links":{"labels":"/api/v2/buckets/03b00a761e9e3001/labels","logs":"/api/v2/buckets/03b00a761e9e3001/logs","members":"/api/v2/buckets/03b00a761e9e3001/members","org":"/api/v2/orgs/03b00a761e9e3000","owners":"/api/v2/buckets/03b00a761e9e3001/owners","self":"/api/v2/buckets/03b00a761e9e3001","write":"/api/v2/write?org=03b00a761e9e3000\u0026bucket=03b00a761e9e3001"},"labels":[]},"org":{"links":{"buckets":"/api/v2/buckets?org={{ .org }}","dashboards":"/api/v2/dashboards?org={{ .org }}","labels":"/api/v2/orgs/03b00a761e9e3000/labels","logs":"/api/v2/orgs/03b00a761e9e3000/logs","members":"/api/v2/orgs/03b00a761e9e3000/members","owners":"/api/v2/orgs/03b00a761e9e3000/owners","secrets":"/api/v2/orgs/03b00a761e9e3000/secrets","self":"/api/v2/orgs/03b00a761e9e3000","tasks":"/api/v2/tasks?org={{ .org }}"},"id":"03b00a761e9e3000","name":"{{ .org }}"},"auth":{"id":"03b00a761e9e3002","token":"d7odFhI50cR8WcLrbfD1pkVenWy51zEM6WC2Md5McGGTxRbOEi5KS0qrXrTEweiH2z5uQjkNa-0YVmpTQlwM3w==","status":"active","description":"{{ .username }}'s Token","orgID":"03b00a761e9e3000","org":"{{ .org }}","userID":"03b00a760bde3000","user":"{{ .username }}","permissions":[{"action":"read","resource":{"type":"authorizations"}},{"action":"write","resource":{"type":"authorizations"}},{"action":"read","resource":{"type":"buckets"}},{"action":"write","resource":{"type":"buckets"}},{"action":"read","resource":{"type":"dashboards"}},{"action":"write","resource":{"type":"dashboards"}},{"action":"read","resource":{"type":"orgs"}},{"action":"write","resource":{"type":"orgs"}},{"action":"read","resource":{"type":"sources"}},{"action":"write","resource":{"type":"sources"}},{"action":"read","resource":{"type":"tasks"}},{"action":"write","resource":{"type":"tasks"}},{"action":"read","resource":{"type":"telegrafs"}},{"action":"write","resource":{"type":"telegrafs"}},{"action":"read","resource":{"type":"users"}},{"action":"write","resource":{"type":"users"}},{"action":"read","resource":{"type":"variables"}},{"action":"write","resource":{"type":"variables"}},{"action":"read","resource":{"type":"scrapers"}},{"action":"write","resource":{"type":"scrapers"}},{"action":"read","resource":{"type":"secrets"}},{"action":"write","resource":{"type":"secrets"}},{"action":"read","resource":{"type":"labels"}},{"action":"write","resource":{"type":"labels"}},{"action":"read","resource":{"type":"views"}},{"action":"write","resource":{"type":"views"}},{"action":"read","resource":{"type":"documents"}},{"action":"write","resource":{"type":"documents"}}],"links":{"self":"/api/v2/authorizations/03b00a761e9e3002","user":"/api/v2/users/03b00a760bde3000"}}}`)
	if err != nil {
		log.Fatal(err)
	}
	req := influxdb.SetupRequest{}
	w.WriteHeader(http.StatusCreated)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Fatal(err)
	}
	if err := resTemplate.Execute(w, map[string]string{"username": req.Username, "org": req.Org, "retentionSeconds": strconv.Itoa(req.RetentionPeriodHrs * 60 * 60)}); err != nil {
		log.Fatal(err)
	}
}

func setupMockServer() (*http.Client, string, func()) {
	sm := http.NewServeMux()
	sm.HandleFunc("/api/v2/write", writeHandler)
	sm.HandleFunc("/api/v2/setup", setupHandler)
	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Fatal(r.RequestURI) // This is usefull for debugging
	})
	server := httptest.NewServer(http.HandlerFunc(sm.ServeHTTP))
	return server.Client(), server.URL, server.Close
}

func setupTLSMockserver() (*http.Client, string, string, string, func()) {
	keyPem := []byte(`-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAomQx2HLKfGpb5SEhNzxIgG/EJw2pElptgm0T4r/pKJOmzSFq
PpPU0qshHUnMqYZuvs8eTGqOSDqo9XXOETwRxHpUPmzrrbgl23HKrOatJCLAcv1N
rNx6NaGjiykTkIsBlmD3TGfH+N0wmtHiB0jjRy8sZHaFUfAUaPOT4ei/i/M/1Muk
06AXbEZ/3+c2fNm+me0hKNJq4JpD768jromhW2D3QjIKFFGroogmd7/9ZIcfvIMb
xbr2KporDK/mLdLrLGrTxjfbtYv+Fuon0ARgmZrcZdwUaFxQXSQjNRe57dr+jTZi
vl8tNFW4uCGiwfox7/vEUz0kV/LC4WssNslwcQIDAQABAoIBAA3wv/6uzAcmMkFX
OLy/JhIwhgw8NflnXeNGbeCXTPK4yibt6Wr50dlL64nSHgmnirZCnX094Hz+3CZG
OKxuFbBiN/0r6Id/OXC/MgDpxI9HlHHKoPJn8u3LtHhrzEwqQragGFqsxhPtGRER
V2/8p9YijJMLQaKpE3d3AYjxLBBdbGlCH3QvS9PwEzqS0fgvaEgKwD13Lu8ZIG2p
RuLVhFMZ3yw44R9PZ29m54EO/qDe4AZNmsZJg+w6ITV7MoodMFcIRkOnNp8pb9no
ZJFjWvkkZKQi7e7u4hSdpqCkqqCH2ml+/5H7PHwOZu8jy5oo/RnDUrsz9Xj/asKr
W8+niUkCgYEAzc8RR6eYPHT56mPMR/EGuwgSuopYkPCMrT9PikwX8njGNrEH9C0V
ikMl2eAXiEO6BaCmw1wXrEyqwgO23NyYQS/PNLxOkqpuY+Ls1H4TQhXimHHyQNcO
Dmis4oudn4xbjiN5/uLtnFhLNatzWKWUxnUkn/dIReXr+srefwFkrRMCgYEAyf6D
owz2p5UbqRzR1EprVjpwuuF7j+/eSXrgEyVFCZcPgCFoFysxREhAxIpB6HBc8Ro9
TwtAWvzf9wWtl2i4mIDnGOYrbyhqurjY5sKFFB4A/TeM5/WOpM9wR5CqGLAyORxo
V8zBRBUnCvFBOwGFwrhupQSPLzca56nKuInJMOsCgYEAonwRm23ArjJ4QMoLtNyg
wLbN+oJRDBUuK3Vpebk7ys35R6KasfeKIv+CebIHQiieS+Ua4+/oLLrWsZg3HcX3
WrfBMlRdAEQYJTo6WkUzNSCMJmkHppNi4JNZsv4hMp6gheaSYV6N07qNnlC/H0SS
4eAIS1bys2Sj2vuhj8nszwsCgYBoGjPdpKC6Xa6TybaaooAPQK84oVz9IbJ+TEWP
mHWsK55hetYamrgZaON4Z4jwMni0CcHvKu1P92O1+8crcV0xu71ep8Fa2ImpEfs3
cqkDZTM9TZPhODz7060aNQR1FNnNdUaReYVhgUVN7mif8Hjvkf30LhVdUBkdq/Q+
h0SZYQKBgQCMCNldHBO0Fj2+gsadmaOnsReCDsGTbxwjdC5psKCWUkOOBCjLXBN9
Mdnk42tqFy88udJphkt5ivBDg9BRldFDgDG0YlIgwfSzJ/7McEspFkjvYbbYctkT
Y+5yAlSc+gJXWkRXzQVA29GxFxmfAKRQrlmw1LrZe6fAxqRQcT0rfQ==
-----END RSA PRIVATE KEY-----`)
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIDITCCAgmgAwIBAgIQQjtyVYP+gdHiKHuD8DS6nDANBgkqhkiG9w0BAQsFADAo
MRMwEQYDVQQKEwppbmZsdXhkYXRhMREwDwYDVQQDEwhpbmZsdXhkYjAgFw0xOTA0
MTIxNDI5MTBaGA8yMDgxMTAxNTIxNDEzN1owKDETMBEGA1UEChMKaW5mbHV4ZGF0
YTERMA8GA1UEAxMIaW5mbHV4ZGIwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQCiZDHYcsp8alvlISE3PEiAb8QnDakSWm2CbRPiv+kok6bNIWo+k9TSqyEd
Scyphm6+zx5Mao5IOqj1dc4RPBHEelQ+bOutuCXbccqs5q0kIsBy/U2s3Ho1oaOL
KROQiwGWYPdMZ8f43TCa0eIHSONHLyxkdoVR8BRo85Ph6L+L8z/Uy6TToBdsRn/f
5zZ82b6Z7SEo0mrgmkPvryOuiaFbYPdCMgoUUauiiCZ3v/1khx+8gxvFuvYqmisM
r+Yt0ussatPGN9u1i/4W6ifQBGCZmtxl3BRoXFBdJCM1F7nt2v6NNmK+Xy00Vbi4
IaLB+jHv+8RTPSRX8sLhayw2yXBxAgMBAAGjRTBDMA4GA1UdDwEB/wQEAwIB5jAP
BgNVHSUECDAGBgRVHSUAMA8GA1UdEwEB/wQFMAMBAf8wDwYDVR0RBAgwBocEfwAA
ATANBgkqhkiG9w0BAQsFAAOCAQEALDtbwwaMqUtN3pocuai/M5ZKi2zDJuEMMhl0
PORlebvAz6voUp9ufdHaxrZMACn2zs/lkRsKl8HEy2ucusOgmW3WkHf7/6TNI6Wz
u/CryeCFIkesdm8BuSX/YLRTzktoYO5xdhv1v0DNYK4fF4W7CKR0Ln2P5/KNNzgS
wxdXhdWyoZKbNI2mS65BqDRPRA/sBPp652969hmGJxk+ZYcefWGX7WgjUdWvFMV6
iN3MeSe0Jnsa1HijHikwz2Z30VqWU2f04jwISsm0Lw2UPFGD67lNi+XgonS7BX73
13BILrvydRraUD2OHlSR3TbXH2Jcdgh7Ifl+Fc0OEaXnlNHT9w==
-----END CERTIFICATE-----`)
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		panic(err)
		//log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(certPem)
	cfg := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool, // We have to set the rootca here becaause we use a self-signed cert, if you use one of the big root cert companies you don't need to use this.
	}
	cfg.BuildNameToCertificate()
	server := httptest.NewUnstartedServer(http.HandlerFunc(writeHandler))
	server.TLS = cfg

	// read in your cert files
	fCert, err := ioutil.TempFile("", "influxdb_example_cert_*.pem")
	if err != nil {
		log.Fatal(err)
	}

	fKey, err := ioutil.TempFile("", "influxdb_example_key_*.pem")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		fCert.Close()
		fKey.Close()
	}()
	// we go to a file here to make the examples more obvious.
	_, err = fCert.Write(certPem)
	if err != nil {
		log.Fatal(err)
	}
	_, err = fKey.Write(keyPem)
	if err != nil {
		log.Fatal(err)
	}

	server.StartTLS()
	return server.Client(), fCert.Name(), fKey.Name(), server.URL, func() { server.Close(); os.Remove(fCert.Name()); os.Remove(fKey.Name()) }
}

// ExampleClient_Write_basic is an example of basic writing to influxdb over http(s).
// While this is fine in a VPN or VPC, we recommend using TLS/HTTPS if you are sending data over the internet, or anywhere
// your tokens could be intercepted.
func ExampleClient_Write_basic() {
	// just us setting up the server so the example will work.  You will likely have to use the old fasioned way to get an *http.Client and address
	// alternatively you can leave the *http.Client nil, and it will intelligently create one with sane defaults.
	myHTTPClient, myHTTPInfluxAddress, teardown := setupMockServer()
	defer teardown() // we shut down our server at the end of the test, obviously you won't be doing this.
	influx, err := influxdb.New(myHTTPClient, influxdb.WithAddress(myHTTPInfluxAddress), influxdb.WithToken("mytoken"))
	if err != nil {
		panic(err) // error handling here, normally we wouldn't use fmt, but it works for the example
	}

	// we use client.NewRowMetric for the example because its easy, but if you need extra performance
	// it is fine to manually build the []client.Metric{}.
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)),
	}

	// The actual write..., this method can be called concurrently.
	if err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		log.Fatal(err) // as above use your own error handling here.
	}
	influx.Close() // closes the client.  After this the client is useless.
	// Output:
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000008
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000009
}

func ExampleClient_Write_tlsMutualAuthentication() {
	// just us setting up the server so the example will work.  You will likely have to use the old fasioned way to get an *http.Client and address
	_, certFileName, keyfileName, myHTTPInfluxAddress, teardown := setupTLSMockserver()
	defer teardown() // we shut down our server at the end of the test, obviously you won't be doing this.

	certPem, err := ioutil.ReadFile(certFileName)
	if err != nil {
		log.Fatal(err)
	}
	keyPem, err := ioutil.ReadFile(keyfileName)
	if err != nil {
		log.Fatal(err)
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()

	// read in the ca cert, in our case since we are self-signing, we are using the same cert
	caCertPem, err := ioutil.ReadFile(certFileName)
	if err != nil {
		log.Fatal(err)
	}
	certPool.AppendCertsFromPEM(caCertPem)

	if err != nil {
		log.Fatal(err)
	}
	tlsConfig := &tls.Config{
		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		// Force it server side
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
	}
	tlsConfig.BuildNameToCertificate()

	influx, err := influxdb.New(influxdb.HTTPClientWithTLSConfig(tlsConfig), influxdb.WithAddress(myHTTPInfluxAddress), influxdb.WithToken("mytoken"))
	if err != nil {
		log.Fatal(err)
	}

	// we use client.NewRowMetric for the example because its easy, but if you need extra performance
	// it is fine to manually build the []client.Metric{}
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)),
	}

	// The actual write...
	if err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		log.Fatal(err)
	}
	influx.Close() // close the client after this the client is useless.
	// Output:
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000008
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000009
}

func ExampleClient_Setup() {
	// just us setting up the server so the example will work.  You will likely have to use the old fasioned way to get an *http.Client and address
	// alternatively you can leave the *http.Client nil, and it will intelligently create one with sane defaults.
	myHTTPClient, myHTTPInfluxAddress, teardown := setupMockServer()
	defer teardown() // we shut down our server at the end of the test, obviously you won't be doing this.

	influx, err := influxdb.New(myHTTPClient, influxdb.WithAddress(myHTTPInfluxAddress), influxdb.WithUserAndPass("my-username", "my-password"))
	if err != nil {
		panic(err) // error handling here, normally we wouldn't use fmt, but it works for the example
	}
	resp, err := influx.Setup(context.Background(), "my-bucket", "my-org", 32)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Auth.Token)
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
	}

	// We can now do a write even though we didn't put a token in
	if err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		log.Fatal(err)
	}
	influx.Close() // close the client after this the client is useless.
	// Output:
	// d7odFhI50cR8WcLrbfD1pkVenWy51zEM6WC2Md5McGGTxRbOEi5KS0qrXrTEweiH2z5uQjkNa-0YVmpTQlwM3w==
	// system-metrics,hostname=hal9000 cpu=0.93,memory=1000i 1520139967000000008
}
