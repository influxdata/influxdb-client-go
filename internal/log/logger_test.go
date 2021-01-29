// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log_test

import (
	"log"
	"strings"
	"testing"

	ilog "github.com/influxdata/influxdb-client-go/v2/internal/log"
	dlog "github.com/influxdata/influxdb-client-go/v2/log"
	"github.com/stretchr/testify/assert"
)

func TestLogging(t *testing.T) {
	var sb strings.Builder
	log.SetOutput(&sb)
	dlog.Log.SetLogLevel(dlog.DebugLevel)
	//test default settings
	ilog.Debug("Debug")
	ilog.Debugf("Debugf %s %d", "message", 1)
	ilog.Info("Info")
	ilog.Infof("Infof %s %d", "message", 2)
	ilog.Warn("Warn")
	ilog.Warnf("Warnf %s %d", "message", 3)
	ilog.Error("Error")
	ilog.Errorf("Errorf %s %d", "message", 4)
	assert.True(t, strings.Contains(sb.String(), "Debug"))
	assert.True(t, strings.Contains(sb.String(), "Debugf message 1"))
	assert.True(t, strings.Contains(sb.String(), "Info"))
	assert.True(t, strings.Contains(sb.String(), "Infof message 2"))
	assert.True(t, strings.Contains(sb.String(), "Warn"))
	assert.True(t, strings.Contains(sb.String(), "Warnf message 3"))
	assert.True(t, strings.Contains(sb.String(), "Error"))
	assert.True(t, strings.Contains(sb.String(), "Errorf message 4"))

	sb.Reset()

	dlog.Log = nil
	ilog.Debug("Debug")
	ilog.Debugf("Debugf %s %d", "message", 1)
	ilog.Info("Info")
	ilog.Infof("Infof %s %d", "message", 2)
	ilog.Warn("Warn")
	ilog.Warnf("Warnf %s %d", "message", 3)
	ilog.Error("Error")
	ilog.Errorf("Errorf %s %d", "message", 4)
	assert.True(t, len(sb.String()) == 0)
}
