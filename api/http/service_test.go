// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestService(t *testing.T) {
	srv := NewService("http://localhost:8086/aa/", "Token my-token", DefaultOptions())
	assert.Equal(t, "http://localhost:8086/aa/", srv.ServerURL())
	assert.Equal(t, "http://localhost:8086/aa/api/v2/", srv.ServerAPIURL())
	assert.Equal(t, "Token my-token", srv.Authorization())
}
