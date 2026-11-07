// Copyright 2026 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTaskReqDetailedEscapesStrings(t *testing.T) {
	every := "1h"
	cron := `0 \ "x"`
	flux := `from(bucket: "b") |> range(start: -1h)`

	req := createTaskReqDetailed(`sensor "east" ${x} \`, flux, &every, nil, "org")
	assert.Equal(t, `option task = { name: "sensor \"east\" \${x} \\", every: 1h } `+flux, req.Flux)

	req = createTaskReqDetailed("plain", flux, nil, &cron, "org")
	assert.Equal(t, `option task = { name: "plain", cron: "0 \\ \"x\"" } `+flux, req.Flux)

	req = createTaskReqDetailed("plain", flux, &every, nil, "org")
	assert.Equal(t, `option task = { name: "plain", every: 1h } `+flux, req.Flux)
}
