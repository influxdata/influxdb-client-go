package gzip_test

import (
	"bytes"
	egzip "compress/gzip"
	"io/ioutil"
	"testing"

	gzip "github.com/influxdata/influxdb-client-go/internal/gzip"
)

func TestGzip(t *testing.T) {
	buf := &bytes.Buffer{}
	r, err := gzip.CompressWithGzip(buf, 4)
	if err != nil {
		t.Fatal(err)
	}

	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	_, err = buf.WriteString(text)
	if err != nil {
		t.Fatal(err)
	}
	ur, err := egzip.NewReader(r)
	if err != nil {
		t.Fatal(err)
	}
	res, err := ioutil.ReadAll(ur)
	if err != nil {
		t.Fatal(err)
	}
	if string(res) != text {
		t.Fatal("text did not encode or possibly decode properly")
	}
}
