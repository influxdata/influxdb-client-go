package influxdb_test

import (
	"compress/gzip"
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	influxdb "github.com/influxdata/influxdb-client-go"
)

func basicHandler(w http.ResponseWriter, r *http.Request) {
	reader := r.Body
	if r.Header.Get("Content-Encoding") == "gzip" {
		var err error
		reader, err = gzip.NewReader(reader)
		if err != nil {
			fmt.Println(err)
		}
	}
	buf, err := ioutil.ReadAll(reader)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(buf))
	w.WriteHeader(200)
}

func setupMockServer() (*http.Client, string, func()) {
	server := httptest.NewServer(http.HandlerFunc(basicHandler))
	return server.Client(), server.URL, server.Close
}

func setupTLSMockserver() (*http.Client, string, string, string, func()) {
	certPem := []byte(`-----BEGIN CERTIFICATE-----
MIIBhTCCASugAwIBAgIQIRi6zePL6mKjOipn+dNuaTAKBggqhkjOPQQDAjASMRAw
DgYDVQQKEwdBY21lIENvMB4XDTE3MTAyMDE5NDMwNloXDTE4MTAyMDE5NDMwNlow
EjEQMA4GA1UEChMHQWNtZSBDbzBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABD0d
7VNhbWvZLWPuj/RtHFjvtJBEwOkhbN/BnnE8rnZR8+sbwnc/KhCk3FhnpHZnQz7B
5aETbbIgmuvewdjvSBSjYzBhMA4GA1UdDwEB/wQEAwICpDATBgNVHSUEDDAKBggr
BgEFBQcDATAPBgNVHRMBAf8EBTADAQH/MCkGA1UdEQQiMCCCDmxvY2FsaG9zdDo1
NDUzgg4xMjcuMC4wLjE6NTQ1MzAKBggqhkjOPQQDAgNIADBFAiEA2zpJEPQyz6/l
Wf86aX6PepsntZv2GYlA5UpabfT2EZICICpJ5h/iI+i341gBmLiAFQOyTDT+/wQc
6MF9+Yw1Yy0t
-----END CERTIFICATE-----`)
	keyPem := []byte(`-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIrYSSNQFaA2Hwf1duRSxKtLYX5CB04fSeQ6tF1aY/PuoAoGCCqGSM49
AwEHoUQDQgAEPR3tU2Fta9ktY+6P9G0cWO+0kETA6SFs38GecTyudlHz6xvCdz8q
EKTcWGekdmdDPsHloRNtsiCa697B2O9IFA==
-----END EC PRIVATE KEY-----`)
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		log.Fatal(err)
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	server := httptest.NewUnstartedServer(http.HandlerFunc(basicHandler))
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
		panic("foo") // error handling here, normally we wouldn't use fmt, but it works for the example
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
		panic(err) // as above use your own error handling here.
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
	tlsConfig := &tls.Config{
		// Reject any TLS certificate that cannot be validated
		ClientAuth: tls.RequireAndVerifyClientCert,
		// Ensure that we only use our "CA" to validate certificates
		// Force it server side
		PreferServerCipherSuites: true,
		Certificates:             []tls.Certificate{cert},
		// TLS 1.2 because we can
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: true,
	}

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
