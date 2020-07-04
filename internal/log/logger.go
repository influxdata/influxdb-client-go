// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log

import (
	"github.com/influxdata/influxdb-client-go/api/log"
)

var Log LogAdapter

// LogAdapter provides filtered and categorized logging API.
// It logs using the api logger instance. Only errors are logged by default.
type LogAdapter struct {
	debugLevel uint
}

// SetDebugLevel to filter log messages. Each level mean to log all categories bellow
// 0 errors , 1 - warning, 2 - info, 3 - debug

func (l *LogAdapter) SetDebugLevel(debugLevel uint) {
	l.debugLevel = debugLevel
}

func (l *LogAdapter) Debugf(format string, v ...interface{}) {
	if log.Log != nil && l.debugLevel > 2 {
		log.Log.Debugf(format, v...)
	}
}

func (l *LogAdapter) Debug(msg string) {
	if log.Log != nil && l.debugLevel > 2 {
		log.Log.Debug(msg)
	}
}

func (l *LogAdapter) Infof(format string, v ...interface{}) {
	if log.Log != nil && l.debugLevel > 1 {
		log.Log.Infof(format, v...)
	}
}

func (l *LogAdapter) Info(msg string) {
	if log.Log != nil && l.debugLevel > 1 {
		log.Log.Info(msg)
	}
}

func (l *LogAdapter) Warnf(format string, v ...interface{}) {
	if log.Log != nil && l.debugLevel > 0 {
		log.Log.Warnf(format, v...)
	}
}

func (l *LogAdapter) Warn(msg string) {
	if log.Log != nil && l.debugLevel > 0 {
		log.Log.Warn(msg)
	}
}

func (l *LogAdapter) Errorf(format string, v ...interface{}) {
	if log.Log != nil {
		log.Log.Errorf(format, v...)
	}
}

func (l *LogAdapter) Error(msg string) {
	if log.Log != nil {
		log.Log.Error(msg)
	}
}
