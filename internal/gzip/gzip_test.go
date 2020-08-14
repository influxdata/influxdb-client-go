// Copyright 2020 InfluxData, Inc.. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package gzip_test

import (
	"bytes"
	egzip "compress/gzip"
	"io/ioutil"
	"testing"

	"github.com/influxdata/influxdb-client-go/v2/internal/gzip"
)

func TestGzip(t *testing.T) {
	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.`
	inputBuffer := bytes.NewBuffer([]byte(text))
	r, err := gzip.CompressWithGzip(inputBuffer)
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
