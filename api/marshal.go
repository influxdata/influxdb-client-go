package api

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	influxdbTag   = "influxdb"
	protoTag      = "protobuf"
	protoOneOfTag = "protobuf_oneof"
)

func MarshalTag(arg any) string {
	log.SetFlags(log.Lshortfile)

	argType := reflect.TypeOf(arg)
	val := reflect.ValueOf(arg)

	var influxPoint string

	numFields := val.NumField()
	count := 0

	for i := 0; i < numFields; i++ {
		fieldVal := val.Field(i)
		fieldName := argType.Field(i).Tag.Get(influxdbTag)

		updatedInfluxPoint := typeHandler(fieldName, fieldVal, nil)
		// add a "," at the end of the string unless the loop is at the last value
		if count < numFields-1 {
			updatedInfluxPoint += ","
		}
		influxPoint += updatedInfluxPoint
		count++
	}
	return influxPoint
}

func MarshalProto(arg any) {
	log.SetFlags(log.Lshortfile)
	var influxPoint string

	//argType := reflect.TypeOf(arg)
	val := reflect.ValueOf(arg)
	argType := reflect.TypeOf(arg)

	numFields := val.NumField()
	count := 0

	for i := 0; i < numFields; i++ {
		fieldVal := reflect.Indirect(val).Field(i)
		fieldProtoTagName := argType.Field(i).Tag.Get(protoTag)
		fieldProtoOneOfTag := argType.Field(i).Tag.Get(protoOneOfTag)
		protoTagNames := strings.Split(fieldProtoTagName, ",")

		if len(protoTagNames) > 1 {
			influxdbTagName := strings.Split(protoTagNames[3], "=")
			updatedInfluxPoint := typeHandler(influxdbTagName[1], fieldVal, fieldVal)
			// add a "," at the end of the string unless the loop is at the last value
			if count < numFields-1 {
				updatedInfluxPoint += ","
			}
			influxPoint += updatedInfluxPoint
		} else {
			updatedInfluxPoint := typeHandler(fieldProtoOneOfTag, fieldVal, fieldVal)
			// add a "," at the end of the string unless the loop is at the last value
			if count < numFields-1 {
				updatedInfluxPoint += ","
			}
			influxPoint += updatedInfluxPoint
		}
		count++
	}
	log.Println(influxPoint)
}

func typeHandler(fieldName string, fieldVal reflect.Value, val any) string {
	spaces := regexp.MustCompile(`\s+`)

	switch fieldVal.Type().String() {
	case "string":
		lowerVal := strings.ToLower(fieldVal.String())
		influxStringVal := spaces.ReplaceAllString(lowerVal, "_")
		return fmt.Sprintf("%v=%v", fieldName, influxStringVal)

	case "time.Time":
		return fmt.Sprintf("%v=%v", fieldName, fieldVal.Interface().(time.Time).Unix())

	default:
		return fmt.Sprintf("%v=%v", fieldName, val)
	}
}
