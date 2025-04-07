package influxclient

import (
	"fmt"
	"reflect"
	"time"
)

// checkContainerType validates the value is struct with simple type fields
// or a map with key as string and value as a simple type
func checkContainerType(p interface{}, alsoMap bool, usage string) error {
	if p == nil {
		return nil
	}
	t := reflect.TypeOf(p)
	v := reflect.ValueOf(p)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() != reflect.Struct && (!alsoMap || t.Kind() != reflect.Map) {
		return fmt.Errorf("cannot use %v as %s", t, usage)
	}
	switch t.Kind() {
	case reflect.Struct:
		fields := reflect.VisibleFields(t)
		for _, f := range fields {
			fv := v.FieldByIndex(f.Index)
			t := getFieldType(fv)
			if !validFieldType(t) {
				return fmt.Errorf("cannot use field '%s' of type '%v' as a %s", f.Name, t, usage)
			}

		}
	case reflect.Map:
		key := t.Key()
		if key.Kind() != reflect.String {
			return fmt.Errorf("cannot use map key of type '%v' for %s name", key, usage)
		}
		for _, k := range v.MapKeys() {
			f := v.MapIndex(k)
			t := getFieldType(f)
			if !validFieldType(t) {
				return fmt.Errorf("cannot use map value type '%v' as a %s", t, usage)
			}
		}
	}
	return nil
}

// getFieldType extracts type of value
func getFieldType(v reflect.Value) reflect.Type {
	t := v.Type()
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Interface && !v.IsNil() {
		t = reflect.ValueOf(v.Interface()).Type()
	}
	return t
}

// timeType is the exact type for the Time
var timeType = reflect.TypeOf(time.Time{})

// validFieldType validates that t is primitive type or string or interface
func validFieldType(t reflect.Type) bool {
	return (t.Kind() > reflect.Invalid && t.Kind() < reflect.Complex64) ||
		t.Kind() == reflect.String ||
		t == timeType
}
