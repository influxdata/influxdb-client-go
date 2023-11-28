package api

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// DataToPoint converts custom point structures into a Point.
// Each visible field of the point on input must be annotated with
// 'lp' prefix and values measurement,tag, field or timestamp.
// Valid point must contain measurement and at least one field.
//
// A field with timestamp must be of a type time.Time
//
//	 type TemperatureSensor struct {
//		  Measurement string `lp:"measurement"`
//		  Sensor string `lp:"tag,sensor"`
//		  ID string `lp:"tag,device_id"`
//		  Temp float64 `lp:"field,temperature"`
//		  Hum int	`lp:"field,humidity"`
//		  Time time.Time `lp:"timestamp,temperature"`
//		  Description string `lp:"-"`
//	 }
func DataToPoint(x interface{}) (*write.Point, error) {
	if err := checkContainerType(x, false, "point"); err != nil {
		return nil, err
	}
	t := reflect.TypeOf(x)
	v := reflect.ValueOf(x)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := reflect.VisibleFields(t)

	var measurement string = ""
	var lpTags = make(map[string]string)
	var lpFields = make(map[string]interface{})
	var lpTime time.Time

	for _, f := range fields {
		name := f.Name
		if tag, ok := f.Tag.Lookup("lp"); ok {
			if tag == "-" {
				continue
			}
			parts := strings.Split(tag, ",")
			if len(parts) > 2 {
				return nil, fmt.Errorf("multiple tag attributes are not supported")
			}
			typ := parts[0]
			if len(parts) == 2 {
				name = parts[1]
			}
			switch typ {
			case "measurement":
				if measurement != "" {
					return nil, fmt.Errorf("multiple measurement fields")
				}
				measurement = v.FieldByIndex(f.Index).String()
			case "tag":
				lpTags[name] = v.FieldByIndex(f.Index).String()
			case "field":
				lpFields[name] = v.FieldByIndex(f.Index).Interface()
			case "timestamp":
				if f.Type != timeType {
					return nil, fmt.Errorf("cannot use field '%s' as a timestamp", f.Name)
				}
				lpTime = v.FieldByIndex(f.Index).Interface().(time.Time)
			default:
				return nil, fmt.Errorf("invalid tag %s", typ)
			}
		}
	}
	if measurement == "" {
		return nil, fmt.Errorf("no struct field with tag 'measurement'")
	}
	if len(lpFields) == 0 {
		return nil, fmt.Errorf("no struct field with tag 'field'")
	}
	return write.NewPoint(measurement, lpTags, lpFields, lpTime), nil
}
