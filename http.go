package client

import (
	"net"
	"net/http"
	"time"
)

func netTransport() *http.Transport {
	return &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   time.Second * 20,
		Transport: netTransport(),
	}
}
