package api

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type goodInfluxTestType struct {
	Measurement string `influxdb:"measurement"`
	Name        string `influxdb:"name"`
	Title       string `influxdb:"title,tag"`
	Distance    int64  `influxdb:"distance"`
	Description string `influxdb:"Description"`
}

type badInfluxTestType2ndArgNotTag struct {
	Measurement string `influxdb:"name,measurement"`
	Name        string `influxdb:"name"`
	Title       string `influxdb:"title,tag"`
	Distance    int64  `influxdb:"distance"`
	Description string `influxdb:"Description"`
}

type badInfluxTestTypeTooManyMeasurements struct {
	Measurement string `influxdb:"measurement"`
	Name        string `influxdb:"name"`
	Title       string `influxdb:"title,tag"`
	Distance    int64  `influxdb:"distance"`
	Description string `influxdb:"measurement"`
}

type badInfluxTestTypeNoMeasurements struct {
	Measurement string `influxdb:"none"`
	Name        string `influxdb:"name"`
	Title       string `influxdb:"title,tag"`
	Distance    int64  `influxdb:"distance"`
	Description string `influxdb:"Description"`
}

var (
	goodInfluxArg = goodInfluxTestType{
		Measurement: "foo",
		Name:        "bar",
		Title:       "test of the struct write point marshaller",
		Distance:    39,
		Description: "This tests the MarshalStructToWritePoint",
	}

	badInfluxArg2ndArgNotTag = badInfluxTestType2ndArgNotTag{
		Measurement: "foo",
		Name:        "bar",
		Title:       "test of the struct write point marshaller",
		Distance:    39,
		Description: "This tests the MarshalStructToWritePoint",
	}

	badInfluxArgTooManyMeasurements = badInfluxTestTypeTooManyMeasurements{
		Measurement: "foo",
		Name:        "bar",
		Title:       "test of the struct write point marshaller",
		Distance:    39,
		Description: "This tests the MarshalStructToWritePoint",
	}

	badInfluxArgNoMeasurements = badInfluxTestTypeNoMeasurements{
		Measurement: "foo",
		Name:        "bar",
		Title:       "test of the struct write point marshaller",
		Distance:    39,
		Description: "This tests the MarshalStructToWritePoint",
	}
)

func Test_MarshalStructToWritePoint_Happy_Path(t *testing.T) {
	point, err := MarshalStructToWritePoint(goodInfluxArg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, point)

	assert.Equal(t, 1, len(point.TagList()))
	assert.Equal(t, 3, len(point.FieldList()))
	assert.Equal(t, "foo", point.Name())
}

func Test_MarshalStructToWritePoint_Sad_Path_2nd_Arg_Not_Taf(t *testing.T) {
	_, err := MarshalStructToWritePoint(badInfluxArg2ndArgNotTag, nil)
	assert.Error(t, err)
	assert.Equal(t, secondTagArgPassedButNotTagErrorMsg, err.Error())
}

func Test_MarshalStructToWritePoint_Sad_Path_Too_Many_Measurements(t *testing.T) {
	_, err := MarshalStructToWritePoint(badInfluxArgTooManyMeasurements, nil)
	assert.Error(t, err)
	assert.Equal(t, tooManyMeasurementsErrorMsg, err.Error())
}

func Test_MarshalStructToWritePoint_Sad_Path_No_Measurements(t *testing.T) {
	_, err := MarshalStructToWritePoint(badInfluxArgNoMeasurements, nil)
	assert.Error(t, err)
	assert.Equal(t, noMeasurementPresentErrorMsg, err.Error())
}
