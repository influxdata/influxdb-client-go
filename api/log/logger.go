// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"log"
)

// Log is the package-wide logger
var Log Logger = &logger{}

type Logger interface {
	SetDebugLevel(debugLevel uint)
	Debugf(format string, v ...interface{})
	Debug(msg string)
	Infof(format string, v ...interface{})
	Info(msg string)
	Warnf(format string, v ...interface{})
	Warn(msg string)
	Errorf(format string, v ...interface{})
	Error(msg string)
}

// logger provides filtered and categorized logging API.
// It logs to standard logger, only errors by default
type logger struct {
	debugLevel uint
}

// SetDebugLevel to filter log messages. Each level mean to log all categories below
// 0 errors , 1 - warning, 2 - info, 3 - debug
func (l *logger) SetDebugLevel(debugLevel uint) {
	l.debugLevel = debugLevel
}

func (l *logger) Debugf(format string, v ...interface{}) {
	if l.debugLevel > 2 {
		log.Print("[D]! ", fmt.Sprintf(format, v...))
	}
}
func (l *logger) Debug(msg string) {
	if l.debugLevel > 2 {
		log.Print("[D]! ", msg)
	}
}

func (l *logger) Infof(format string, v ...interface{}) {
	if l.debugLevel > 1 {
		log.Print("[I]! ", fmt.Sprintf(format, v...))
	}
}
func (l *logger) Info(msg string) {
	if l.debugLevel > 1 {
		log.Print("[I]! ", msg)
	}
}

func (l *logger) Warnf(format string, v ...interface{}) {
	if l.debugLevel > 0 {
		log.Print("[W]! ", fmt.Sprintf(format, v...))
	}
}
func (l *logger) Warn(msg string) {
	if l.debugLevel > 0 {
		log.Print("[W]! ", msg)
	}
}

func (l *logger) Errorf(format string, v ...interface{}) {
	log.Print("[E]! ", fmt.Sprintf(format, v...))
}

func (l *logger) Error(msg string) {
	log.Print("[E]! ", msg)
}
