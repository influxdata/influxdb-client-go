// Copyright 2020 InfluxData, Inc. All rights reserved.
// Use of this source code is governed by MIT
// license that can be found in the LICENSE file.

package write

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Point extension methods for test

// ToLineProtocol creates InfluxDB line protocol string from the Point, converting associated timestamp according to precision
// and write result to the string builder
func PointToLineProtocolBuffer(p *Point, sb *strings.Builder, precision time.Duration) {
	escapeKey(sb, p.Name())
	sb.WriteRune(',')
	for i, t := range p.TagList() {
		if i > 0 {
			sb.WriteString(",")
		}
		escapeKey(sb, t.Key)
		sb.WriteString("=")
		escapeKey(sb, t.Value)
	}
	sb.WriteString(" ")
	for i, f := range p.FieldList() {
		if i > 0 {
			sb.WriteString(",")
		}
		escapeKey(sb, f.Key)
		sb.WriteString("=")
		switch f.Value.(type) {
		case string:
			sb.WriteString(`"`)
			escapeValue(sb, f.Value.(string))
			sb.WriteString(`"`)
		default:
			sb.WriteString(fmt.Sprintf("%v", f.Value))
		}
		switch f.Value.(type) {
		case int64:
			sb.WriteString("i")
		case uint64:
			sb.WriteString("u")
		}
	}
	if !p.Time().IsZero() {
		sb.WriteString(" ")
		switch precision {
		case time.Microsecond:
			sb.WriteString(strconv.FormatInt(p.Time().UnixNano()/1000, 10))
		case time.Millisecond:
			sb.WriteString(strconv.FormatInt(p.Time().UnixNano()/1000000, 10))
		case time.Second:
			sb.WriteString(strconv.FormatInt(p.Time().Unix(), 10))
		default:
			sb.WriteString(strconv.FormatInt(p.Time().UnixNano(), 10))
		}
	}
	sb.WriteString("\n")
}

// ToLineProtocol creates InfluxDB line protocol string from the Point, converting associated timestamp according to precision
func PointToLineProtocol(p *Point, precision time.Duration) string {
	var sb strings.Builder
	sb.Grow(1024)
	PointToLineProtocolBuffer(p, &sb, precision)
	return sb.String()
}

func escapeKey(sb *strings.Builder, key string) {
	for _, r := range key {
		switch r {
		case ' ', ',', '=':
			sb.WriteString(`\`)
		}
		sb.WriteRune(r)
	}
}

func escapeValue(sb *strings.Builder, value string) {
	for _, r := range value {
		switch r {
		case '\\', '"':
			sb.WriteString(`\`)
		}
		sb.WriteRune(r)
	}
}
