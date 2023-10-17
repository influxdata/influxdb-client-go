package api

import (
	"errors"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	influxdbTag                         = "influxdb"
	tooManyMeasurementsErrorMsg         = "more than 1 struct field is tagged as a measurement. Please pick only 1 struct field to be a measurement"
	measurementIsNotStringErrorMsg      = "the value for the struct field tagged for measurement is not of type string"
	tagValueNotStringErrorMsg           = "the value for the struct field for a tag is not of type string"
	noMeasurementPresentErrorMsg        = "no struct field is tagged as a measurement. You must have a measurement"
	tooManyTagArgs                      = "your influx tag contains more than the allowed number of arguments"
	secondTagArgPassedButNotTagErrorMsg = "your influx tag has a second argument but it is not for a tag. If you're trying to set a struct field to be a measurement than the only argument that can be passed is 'measurement'"
)

type Tags map[string]string
type Fields map[string]interface{}

func MarshalStructToWritePoint(arg interface{}, timestamp *time.Time) (*write.Point, error) {
	var measurement string
	var tags Tags = make(map[string]string)
	var fields Fields = make(map[string]interface{})

	measurementCount := 0
	ts := time.Now().UTC()

	if timestamp != nil {
		ts = *timestamp
	}
	log.SetFlags(log.Lshortfile)

	argType := reflect.TypeOf(arg)
	val := reflect.ValueOf(arg)

	numFields := val.NumField()

	for i := 0; i < numFields; i++ {
		if measurementCount > 1 {
			return nil, errors.New(tooManyMeasurementsErrorMsg)
		}
		structFieldVal := val.Field(i)
		structFieldName := argType.Field(i).Tag.Get(influxdbTag)

		err := checkEitherTagOrMeasurement(structFieldName)
		if err != nil {
			return nil, err
		}

		if structFieldName == "measurement" {
			measurementFieldVal, ok := structFieldVal.Interface().(string)
			if !ok {
				return nil, errors.New(measurementIsNotStringErrorMsg)
			}
			measurement = measurementFieldVal
			measurementCount++
			continue
		}

		if strings.Contains(structFieldName, "tag") {
			stringTagVal, ok := structFieldVal.Interface().(string)
			if !ok {
				return nil, errors.New(tagValueNotStringErrorMsg)
			}
			tags[structFieldName] = stringTagVal
			continue
		}

		parsedFieldVal := fieldTypeHandler(structFieldVal)
		fields[structFieldName] = parsedFieldVal
	}

	if measurementCount == 0 {
		return nil, errors.New(noMeasurementPresentErrorMsg)
	}

	if measurementCount > 1 {
		return nil, errors.New(tooManyMeasurementsErrorMsg)
	}

	return write.NewPoint(measurement, tags, fields, ts), nil
}

func fieldTypeHandler(fieldVal interface{}) interface{} {
	spaces := regexp.MustCompile(`\s+`)

	switch fieldValType := fieldVal.(type) {
	case string:
		lowerVal := strings.ToLower(fieldValType)
		influxStringVal := spaces.ReplaceAllString(lowerVal, "_")
		return influxStringVal

	case time.Time:
		return fieldValType.Unix()

	default:
		return fieldVal
	}
}

func checkEitherTagOrMeasurement(influxTag string) error {
	tags := strings.Split(influxTag, ",")

	if len(tags) > 2 {
		return errors.New(tooManyTagArgs)
	}

	if len(tags) == 2 && !strings.Contains(tags[1], "tag") {
		return errors.New(secondTagArgPassedButNotTagErrorMsg)
	}

	return nil
}
