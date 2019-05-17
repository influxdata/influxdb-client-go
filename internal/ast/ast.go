// Package ast provides tools for manipulating the flux ast.
// Eventually this will become a builder api.
package ast

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

func FluxExtern(m ...interface{}) (*Extern, error) {
	if len(m) == 0 { // early exit
		return nil, nil
	}
	res := &Extern{Type: "File", Body: make([]variableAssignment, 0, len(m))}
	for i := range m {
		v := reflect.ValueOf(m[i])
		ty := v.Type()
		switch ty.Kind() {
		case reflect.Map:
			if ty.Key().Kind() != reflect.String {
				return nil, &typeConversionError{ty}
			}
			r := v.MapRange()
			for r.Next() {
				val := r.Value()
				switch val.Kind() {
				case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
					if val.IsNil() {
						continue
					}
				}
				if !val.CanInterface() {
					continue
				}
				ival, err := convert(val.Interface())
				if err != nil {
					return nil, err
				}
				name := r.Key().String()
				if err = validateVarName(name); err != nil {
					return nil, err
				}
				stmt := variableAssignment{
					Name: r.Key().String(),
					Init: ival,
				}
				res.Body = append(res.Body, stmt)
			}
		case reflect.Struct:
			l := ty.NumField()
		fieldIteration:
			for i := 0; i < l; i++ {
				f := ty.Field(i)
				val := v.Field(i)
				name := f.Name
				// ignore private fields
				if !val.CanInterface() {
					continue
				}
				// flux doesn't support nil values in objects
				switch val.Kind() {
				case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
					if val.IsNil() {
						continue
					}
				}
				// handle tags
				if tag, ok := f.Tag.Lookup("flux"); ok {
					tags := strings.Split(tag, `,`)
					if len(tags) > 0 {
						if tags[0] != "" {
							name = tags[0]
						}
						if len(tags) > 1 && isEmptyValue(val) {
							for j := range tags {
								if tags[j] == "omitempty" {
									continue fieldIteration
								}
							}
						}
					}
				}
				if err := validateVarName(name); err != nil {
					return nil, err
				}
				fieldInterfaceVal, err := convert(val.Interface())
				if err != nil {
					return nil, err
				}
				stmt := variableAssignment{
					Name: name,
					Init: fieldInterfaceVal,
				}
				res.Body = append(res.Body, stmt)
			}

		default:
			return nil, typeError{ty: ty}
		}
	}
	return res, nil
}

type Extern struct {
	Type string               `json:"type"`
	Body []variableAssignment `json:"body"`
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
	switch x := x.(type) {
	case bool:
		return genericLiteral{"BooleanLiteral", x}, nil
	case int:
		return integerLiteral{int64(x)}, nil
	case int16:
		return integerLiteral{int64(x)}, nil
	case int32:
		return integerLiteral{int64(x)}, nil
	case int64:
		return integerLiteral{int64(x)}, nil
	case uint:
		return uintLiteral{uint64(x)}, nil
	case uint16:
		return uintLiteral{uint64(x)}, nil
	case uint32:
		return uintLiteral{uint64(x)}, nil
	case uint64:
		return uintLiteral{uint64(x)}, nil
	case time.Time, *time.Time:
		return genericLiteral{"DateTimeLiteral", x}, nil
	case string:
		return genericLiteral{"StringLiteral", x}, nil
	case time.Duration:
		return parseSignedduration(x.String())
	case *regexp.Regexp:
		return genericLiteral{Type: "RegexpLiteral", Value: x.String()}, nil
	case regexp.Regexp:
		return genericLiteral{Type: "RegexpLiteral", Value: x.String()}, nil
	case nil:
		return nil, nil
	default:
		v := reflect.ValueOf(x)
		switch reflect.TypeOf(x).Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				return nil, nil
			}
			return convert(v.Elem().Interface())
		case reflect.Array:
			if v.IsNil() {
				return nil, nil
			}
			if !v.IsValid() {
				return nil, nil
			}
			return convert(v.Slice(0, v.Len()))
		case reflect.Struct:
			t := reflect.TypeOf(x)
			l := t.NumField()
			res := objectExpression{Properties: make([]property, 0, l)}
			for i := 0; i < l; i++ {
				f := t.Field(i)
				val := v.Field(i)
				name := f.Name
				// check if field is exported
				if !val.CanInterface() {
					continue
				}

				// flux doesn't support nil values in objects
				switch val.Kind() {
				case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
					if val.IsNil() {
						continue
					}
				}
				// handle tags
				if tag, ok := f.Tag.Lookup("flux"); ok {
					tags := strings.Split(tag, `,`)
					if len(tags) > 0 {
						if tags[0] != "" {
							name = tags[0]
						}
						if len(tags) > 1 && isEmptyValue(val) {
							for j := range tags[1:] {
								if tags[j] == "omitempty" {
									continue
								}
							}
						}
					}
				}
				fieldInterfaceVal, err := convert(val.Interface())
				if err != nil {
					return nil, err
				}
				res.Properties = append(res.Properties, property{Key: identifier{Name: name}, Value: fieldInterfaceVal})
			}
			return res, nil
		case reflect.Map:
			t := reflect.TypeOf(x)
			if t.Key().Kind() != reflect.String {
				return nil, &typeConversionError{t}
			}
			r := v.MapRange()
			res := objectExpression{Properties: make([]property, 0, v.Len())}
			for r.Next() {
				if !r.Value().CanInterface() {
					continue
				}
				val, err := convert(r.Value().Interface())
				if err != nil {
					return nil, err
				}
				if val == nil {
					continue
				}
				res.Properties = append(res.Properties, property{Key: stringLiteral{r.Key().String()}, Value: val})
			}
			return res, nil
		case reflect.Slice:
			l := v.Len()
			res := arrayExpression{Elements: make([]interface{}, 0, l)}
			for i := 0; i < l; i++ {
				relem := v.Index(i)
				if relem.IsNil() {
					continue
				}
				elem, err := convert(relem.Interface())
				if err != nil {
					return nil, err
				}
				res.Elements = append(res.Elements, elem)
			}
		case reflect.Interface:
			elem := v.Elem()
			if !elem.CanInterface() {
				return nil, nil
			}
			return convert(elem.Interface())
		}
		return nil, &typeConversionError{reflect.TypeOf(x)}
	}
}

type variableAssignment struct {
	Name string      `json:"name"`
	Init interface{} `json:"init"`
}

func (e *variableAssignment) MarshalJSON() ([]byte, error) {
	type id struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}
	raw := struct {
		Type string      `json:"type"`
		ID   id          `json:"id"`
		Init interface{} `json:"init"`
	}{
		Type: "VariableAssignment",
		ID: id{
			Type: "Identifier",
			Name: e.Name,
		},
		Init: e.Init,
	}
	return json.Marshal(raw)
}

type duration struct {
	Magnitude int64  `json:"magnitude"`
	Unit      string `json:"unit"`
}

type durationLiteral struct {
	Type   string
	Values []duration `json:"values"`
}

type genericLiteral struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type integerLiteral struct {
	Value int64 `json:"value,string"`
}

func (e integerLiteral) MarshalJSON() ([]byte, error) {
	raw := struct {
		Type  string `json:"type"`
		Value int64  `json:"value,string"`
	}{
		Type:  "IntegerLiteral",
		Value: e.Value,
	}
	return json.Marshal(raw)
}

type stringLiteral struct {
	Value string `json:"value"`
}

func (e stringLiteral) key() string {
	return e.Value
}

func (e stringLiteral) MarshalJSON() ([]byte, error) {
	raw := struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  "StringLiteral",
		Value: e.Value,
	}
	return json.Marshal(raw)
}

type uintLiteral struct {
	Value uint64 `json:"value,string"`
}

func (e uintLiteral) MarshalJSON() ([]byte, error) {
	raw := struct {
		Type  string `json:"type"`
		Value uint64 `json:"value,string"`
	}{
		Type:  "UnsignedIntegerLiteral",
		Value: e.Value,
	}
	return json.Marshal(raw)
}

type keyer interface {
	key() string
}

type property struct {
	Key   keyer       `json:"key"`
	Value interface{} `json:"value"`
}

func (e property) MarshalJSON() ([]byte, error) {
	raw := struct {
		Type  string      `json:"type"`
		Key   keyer       `json:"key"`
		Value interface{} `json:"value"`
	}{
		Type:  "Property",
		Key:   e.Key,
		Value: e.Value,
	}
	return json.Marshal(raw)
}

type objectExpression struct {
	Properties []property `json:"properties"`
}

func (e objectExpression) MarshalJSON() ([]byte, error) {
	raw := struct {
		Type       string     `json:"type"`
		Properties []property `json:"properties"`
	}{
		Type:       "ObjectExpression",
		Properties: e.Properties,
	}
	return json.Marshal(raw)
}

// ParseSignedduration will convert a string into a possibly negative durationLiteral.
func parseSignedduration(lit string) (*durationLiteral, error) {
	r, s := utf8.DecodeRuneInString(lit)
	if r == '-' {
		d, err := parseduration(lit[s:])
		if err != nil {
			return nil, err
		}
		for i := range d.Values {
			d.Values[i].Magnitude = -d.Values[i].Magnitude
		}
		return d, nil
	}
	return parseduration(lit)
}

// parseduration will convert a string into an durationLiteral.
func parseduration(lit string) (*durationLiteral, error) {
	// parseduration will convert a string into components of the duration.
	var values []duration
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
		values = append(values, duration{
			Magnitude: magnitude,
			Unit:      unit,
		})
		lit = lit[n:]
	}
	return &durationLiteral{Values: values}, nil
}

type identifier struct {
	Name string `json:"name"`
}

func (x identifier) key() string {
	return x.Name
}

func (x identifier) MarshalJSON() ([]byte, error) {
	a := struct {
		Type string `json:"type"`
		Name string `json:"name"`
	}{
		Type: "Identifier",
		Name: x.Name,
	}
	return json.Marshal(a)
}

type arrayExpression struct {
	Elements []interface{} `json:"elements"`
}

// this function adapted from https://golang.org/src/encoding/json/encode.go
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func validateVarName(s string) error {
	if len(s) == 0 {
		return errors.New("a legal variable name must be longer than 0")
	}
	r, n := utf8.DecodeRuneInString(s)
	if !unicode.IsLetter(r) {
		return fmt.Errorf("variable names must start with a unicode letter, %s does not", s)
	}

	for _, r := range s[n:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return fmt.Errorf("variable names must be a unicode letter optionally followed by unicode letters and digits, %s does not", s)
		}
	}
	return nil
}
