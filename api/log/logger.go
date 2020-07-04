// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"log"
)

// Log is the package-wide logger
var Log Logger = &logLogger{}

// Logger provides a filtered logging implementation.
type Logger interface {
	Debugf(format string, v ...interface{})
	Debug(msg string)
	Infof(format string, v ...interface{})
	Info(msg string)
	Warnf(format string, v ...interface{})
	Warn(msg string)
	Errorf(format string, v ...interface{})
	Error(msg string)
}

// logLogger provides a default logging API implementation using log.
// Handling of log levels is is done in the internal/log adapter
type logLogger struct{}

func (l *logLogger) Debugf(format string, v ...interface{}) {
	log.Print("[D]! ", fmt.Sprintf(format, v...))
}

func (l *logLogger) Debug(msg string) {
	log.Print("[D]! ", msg)
}

func (l *logLogger) Infof(format string, v ...interface{}) {
	log.Print("[I]! ", fmt.Sprintf(format, v...))
}

func (l *logLogger) Info(msg string) {
	log.Print("[I]! ", msg)
}

func (l *logLogger) Warnf(format string, v ...interface{}) {
	log.Print("[W]! ", fmt.Sprintf(format, v...))
}

func (l *logLogger) Warn(msg string) {
	log.Print("[W]! ", msg)
}

func (l *logLogger) Errorf(format string, v ...interface{}) {
	log.Print("[E]! ", fmt.Sprintf(format, v...))
}

func (l *logLogger) Error(msg string) {
	log.Print("[E]! ", msg)
}
