// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package log

import (
	"fmt"
	"log"
)

var Log Logger

// Logger provides filtered and categorized logging API.
// It logs to standard logger, only errors by default
type Logger struct {
	debugLevel uint
}

// SetDebugLevel to filter log messages. Each level mean to log all categories bellow
// 0 errors , 1 - warning, 2 - info, 3 - debug
func (l *Logger) SetDebugLevel(debugLevel uint) {
	l.debugLevel = debugLevel
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if l.debugLevel > 2 {
		log.Print("[D]! ", fmt.Sprintf(format, v...))
	}
}
func (l *Logger) Debug(msg string) {
	if l.debugLevel > 2 {
		log.Print("[D]! ", msg)
	}
}

func (l *Logger) Infof(format string, v ...interface{}) {
	if l.debugLevel > 1 {
		log.Print("[I]! ", fmt.Sprintf(format, v...))
	}
}
func (l *Logger) Info(msg string) {
	if l.debugLevel > 1 {
		log.Print("[I]! ", msg)
	}
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	if l.debugLevel > 0 {
		log.Print("[W]! ", fmt.Sprintf(format, v...))
	}
}
func (l *Logger) Warn(msg string) {
	if l.debugLevel > 0 {
		log.Print("[W]! ", msg)
	}
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	log.Print("[E]! ", fmt.Sprintf(format, v...))
}

func (l *Logger) Error(msg string) {
	log.Print("[E]! ", msg)
}
