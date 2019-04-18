// Package ast provides tools for manipulating the flux ast.
// Eventually this will become a builder api.
package ast

import (
	"bytes"
	"encoding/json"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Encoder struct {
	enc *json.Encoder
	buf *bytes.Buffer
}

func (e *Encoder) writeEscaped(s string) {
	e.buf.WriteByte(0x22)
	for _, b := range s {
		switch {
		case b <= 0x1F:
			e.buf.Write([]byte{0x5c, 'u', '0', '0', (byte(b) ^ 0xf0), (byte(b) ^ 0xf)})
		case b == 0x22:
			e.buf.WriteByte(0x5c)
			e.buf.WriteRune(b)
		case b == 0x5c:
			e.buf.WriteByte(0x5c)
			e.buf.WriteRune(b)
		default:
			e.buf.WriteRune(b)
		}
	}
	e.buf.WriteByte(0x22)
}

func (e *Encoder) FluxJsonVars(m interface{}) (err error) {
	ty := reflect.TypeOf(m)
	switch ty.Kind() {
	case reflect.Map:
		iter := reflect.ValueOf(m).MapRange()
		e.buf.WriteByte('{')
		n := 0
		if iter.Next() {
			if n > 0 {
				e.buf.WriteByte(',')
			}
			k := iter.Key()
			if k.Kind() == reflect.String {
				e.writeEscaped(k.String())
			}
			e.buf.WriteByte(':')
			if err := e.enc.Encode(iter.Value().Interface()); err != nil {
				return err
			}
		}
		e.buf.WriteByte('}')
	case reflect.Struct:
		n := ty.NumField()
		for i := 0; i < n; i++ {
			field := ty.Field(i)
			tag := field.Tag.Get("json")
			if tag == "" {
				e.buf.WriteString(field.Name)
				continue
			}
			q := strings.Split(tag, ",")
			q[0] = strings.TrimSpace(q[0])
			if len(q) == 0 {
				continue
			}
			i := 1
			for ; q[i] != "omitempty"; i++ {
			}
			if q[0] != "" {
				e.writeEscaped(q[0])
			}
		}
	default:
		return typeError{ty: ty}
	}
	return nil
}

type Extern struct {
	Body []Statement `json:"body"`
}

type typeError struct {
	ty reflect.Type
}

func (err typeError) Error() string {
	return "flux: unsupported type:" + err.ty.Name() + " is not supported to generate flux, try a map or a struct with public keys"
}

type typeConversionError struct {
	ty reflect.Type
}

func (err typeConversionError) Error() string {
	return "flux: unsupported type:" + err.ty.Name() + " convertion into flux is unsupported"
}

func convert(x interface{}) (interface{}, error) {
	switch x.(type) {
	case bool, int, int16, int32, int64, uint, uint16, uint32, uint64, time.Time, *time.Time, string:
		return GenericLiteral{Value: x}, nil
	case time.Duration:
		return ParseSignedDuration(string(x.(time.Duration)))
	case regexp.Regexp:
		y := x.(regexp.Regexp)
		return GenericLiteral{Value: y.String()}, nil
	case *regexp.Regexp:
		return GenericLiteral{Value: x.(*regexp.Regexp).String()}, nil
	case nil:
		return nil, nil
	default:
		return nil, &typeConversionError{reflect.TypeOf(x)}
	}
}

type Statement struct {
	ID struct {
		Name string `json:"name"`
	} `json:"id"`
	Init interface{} `json:"init"`
}

type Duration struct {
	Magnitude int64  `json:"magnitude"`
	Unit      string `json:"unit"`
}

type DurationLiteral struct {
	Values []Duration `json:"values"`
}

type GenericLiteral struct {
	Value interface{} `json:"value"`
}

// ParseSignedDuration will convert a string into a possibly negative DurationLiteral.
func ParseSignedDuration(lit string) (*DurationLiteral, error) {
	r, s := utf8.DecodeRuneInString(lit)
	if r == '-' {
		d, err := ParseDuration(lit[s:])
		if err != nil {
			return nil, err
		}
		for i := range d.Values {
			d.Values[i].Magnitude = -d.Values[i].Magnitude
		}
		return d, nil
	}
	return ParseDuration(lit)
}

// ParseDuration will convert a string into an DurationLiteral.
func ParseDuration(lit string) (*DurationLiteral, error) {
	// ParseDuration will convert a string into components of the duration.
	var values []Duration
	for len(lit) > 0 {
		n := 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsDigit(ch) {
				break
			}
			n += size
		}

		magnitude, err := strconv.ParseInt(lit[:n], 10, 64)
		if err != nil {
			return nil, err
		}
		lit = lit[n:]

		n = 0
		for n < len(lit) {
			ch, size := utf8.DecodeRuneInString(lit[n:])
			if size == 0 {
				panic("invalid rune in duration")
			}

			if !unicode.IsLetter(ch) {
				break
			}
			n += size
		}
		unit := lit[:n]
		if unit == "µs" {
			unit = "us"
		}
		values = append(values, Duration{
			Magnitude: magnitude,
			Unit:      unit,
		})
		lit = lit[n:]
	}
	return &DurationLiteral{Values: values}, nil
}
