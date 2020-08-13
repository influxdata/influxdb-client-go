// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log_test

import (
	"fmt"
	"log"
	"strings"
	"testing"

	dlog "github.com/influxdata/influxdb-client-go/log"
	"github.com/stretchr/testify/assert"
)

func logMessages() {
	dlog.Log.Debug("Debug")
	dlog.Log.Debugf("Debugf %s %d", "message", 1)
	dlog.Log.Info("Info")
	dlog.Log.Infof("Infof %s %d", "message", 2)
	dlog.Log.Warn("Warn")
	dlog.Log.Warnf("Warnf %s %d", "message", 3)
	dlog.Log.Error("Error")
	dlog.Log.Errorf("Errorf %s %d", "message", 4)
}

func verifyLogs(t *testing.T, sb *strings.Builder, logLevel uint, prefix string) {
	if logLevel >= dlog.DebugLevel {
		assert.True(t, strings.Contains(sb.String(), prefix+" D! Debug"))
		assert.True(t, strings.Contains(sb.String(), prefix+" D! Debugf message 1"))
	} else {
		assert.False(t, strings.Contains(sb.String(), prefix+" D! Debug"))
		assert.False(t, strings.Contains(sb.String(), prefix+" D! Debugf message 1"))
	}
	if logLevel >= dlog.InfoLevel {
		assert.True(t, strings.Contains(sb.String(), prefix+" I! Info"))
		assert.True(t, strings.Contains(sb.String(), prefix+" I! Infof message 2"))
	} else {
		assert.False(t, strings.Contains(sb.String(), prefix+" I! Info"))
		assert.False(t, strings.Contains(sb.String(), prefix+" I! Infof message 2"))

	}
	if logLevel >= dlog.WarningLevel {
		assert.True(t, strings.Contains(sb.String(), prefix+" W! Warn"))
		assert.True(t, strings.Contains(sb.String(), prefix+" W! Warnf message 3"))
	} else {
		assert.False(t, strings.Contains(sb.String(), prefix+" W! Warn"))
		assert.False(t, strings.Contains(sb.String(), prefix+" W! Warnf message 3"))
	}
	if logLevel >= dlog.ErrorLevel {
		assert.True(t, strings.Contains(sb.String(), prefix+" E! Error"))
		assert.True(t, strings.Contains(sb.String(), prefix+" E! Errorf message 4"))
	}
}

func TestLogging(t *testing.T) {
	var sb strings.Builder
	log.SetOutput(&sb)
	log.SetFlags(0)
	//test default settings
	logMessages()
	verifyLogs(t, &sb, dlog.ErrorLevel, "influxdb2client")

	sb.Reset()
	dlog.Log.SetLogLevel(dlog.WarningLevel)
	logMessages()
	verifyLogs(t, &sb, dlog.WarningLevel, "influxdb2client")

	sb.Reset()
	dlog.Log.SetLogLevel(dlog.InfoLevel)
	logMessages()
	verifyLogs(t, &sb, dlog.InfoLevel, "influxdb2client")

	sb.Reset()
	dlog.Log.SetLogLevel(dlog.DebugLevel)
	logMessages()
	verifyLogs(t, &sb, dlog.DebugLevel, "influxdb2client")

	sb.Reset()
	dlog.Log.SetPrefix("client")
	logMessages()
	verifyLogs(t, &sb, dlog.DebugLevel, "client")
}

func TestCustomLogger(t *testing.T) {
	var sb strings.Builder
	log.SetOutput(&sb)
	log.SetFlags(0)
	dlog.Log = &testLogger{}
	//test default settings
	logMessages()
	verifyLogs(t, &sb, dlog.DebugLevel, "testlogger")
}

type testLogger struct {
}

func (l *testLogger) SetLogLevel(_ uint) {
}

func (l *testLogger) SetPrefix(_ string) {
}

func (l *testLogger) Debugf(format string, v ...interface{}) {
	log.Print("testlogger", " D! ", fmt.Sprintf(format, v...))
}
func (l *testLogger) Debug(msg string) {
	log.Print("testlogger", " D! ", msg)
}

func (l *testLogger) Infof(format string, v ...interface{}) {
	log.Print("testlogger", " I! ", fmt.Sprintf(format, v...))
}
func (l *testLogger) Info(msg string) {
	log.Print("testlogger", " I! ", msg)
}

func (l *testLogger) Warnf(format string, v ...interface{}) {
	log.Print("testlogger", " W! ", fmt.Sprintf(format, v...))
}
func (l *testLogger) Warn(msg string) {
	log.Print("testlogger", " W! ", msg)
}

func (l *testLogger) Errorf(format string, v ...interface{}) {
	log.Print("testlogger", " E! ", fmt.Sprintf(format, v...))
}

func (l *testLogger) Error(msg string) {
	log.Print("testlogger", " [E]! ", msg)
}
