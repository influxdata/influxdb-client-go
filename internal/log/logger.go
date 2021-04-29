// Copyright 2020-2021 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

// Package log provides internal logging infrastructure
package log

import (
	ilog "github.com/influxdata/influxdb-client-go/v2/log"
)

func Debugf(format string, v ...interface{}) {
	if ilog.Log != nil {
		ilog.Log.Debugf(format, v...)
	}
}
func Debug(msg string) {
	if ilog.Log != nil {
		ilog.Log.Debug(msg)
	}
}

func Infof(format string, v ...interface{}) {
	if ilog.Log != nil {
		ilog.Log.Infof(format, v...)
	}
}
func Info(msg string) {
	if ilog.Log != nil {
		ilog.Log.Info(msg)
	}
}

func Warnf(format string, v ...interface{}) {
	if ilog.Log != nil {
		ilog.Log.Warnf(format, v...)
	}
}
func Warn(msg string) {
	if ilog.Log != nil {
		ilog.Log.Warn(msg)
	}
}

func Errorf(format string, v ...interface{}) {
	if ilog.Log != nil {
		ilog.Log.Errorf(format, v...)
	}
}

func Error(msg string) {
	if ilog.Log != nil {
		ilog.Log.Error(msg)
	}
}

func LogLevel() uint {
	if ilog.Log != nil {
		return ilog.Log.LogLevel()
	}
	return ilog.ErrorLevel
}
