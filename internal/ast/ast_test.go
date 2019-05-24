package ast

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/influxdata/flux"
	fluxast "github.com/influxdata/flux/ast"
)

func ExampleFluxExtern_map() {
	buf := &strings.Builder{}
	enc := json.NewEncoder(buf)
	ext, err := FluxExtern(map[string]string{
		"bucket": "my-bucket",
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := enc.Encode(ext); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
	// Output:
	// {"type":"File","body":[{"type":"VariableAssignment","id":{"type":"Identifier","name":"bucket"},"init":{"type":"StringLiteral","value":"my-bucket"}}]}
}

func ExampleFluxExtern_struct() {
	buf := &strings.Builder{}
	enc := json.NewEncoder(buf)
	ext, err := FluxExtern(struct {
		Bucket        string `flux:"bucket"`
		OtherVariable string `flux:"otherVariable,omitempty"`
	}{
		Bucket: "my-bucket",
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := enc.Encode(ext); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
	// Output:
	// {"type":"File","body":[{"type":"VariableAssignment","id":{"type":"Identifier","name":"bucket"},"init":{"type":"StringLiteral","value":"my-bucket"}}]}
}

func TestAst(t *testing.T) {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	res, err := FluxExtern(
		map[string]interface{}{
			"hello":            "x",
			"hellofriend":      3,
			"uinsignedIntTest": uint(7),
			"nowabool":         false,
			"fred": struct {
				A        interface{} `flux:"cow"`
				B        int
				DontShow interface{}
				private  int
			}{
				A:        map[string]bool{"9": true},
				B:        1776,
				DontShow: nil,
				private:  99,
			},
			"empty": nil,
			"complex": map[string]interface{}{
				"hello": struct{}{},
				"okra":  regexp.MustCompile(`[a-z]+`),
				"mapmapmap": map[string]map[string]map[string]int{
					"map0": {"map1": {"map2": 0}},
				},
			},
		},
		struct {
			private        int
			Public         interface{}
			OmitEmptyTest1 string `flux:",omitempty"`
			OmitEmptyTest2 string `flux:"omit_empty_test_snake,omitempty"`
			OmitEmptyTest3 string `flux:",omitempty"`
			OmitEmptyTest4 string `flux:"omit_empty_test_snake2,omitempty"`
		}{
			private: 9,
			Public: map[string]time.Time{
				"perl harbor day": time.Date(1941, time.December, 7, 7, 48, 0, 0, func() *time.Location {
					t, _ := time.LoadLocation("US/Hawaii")
					return t
				}()),
			},
			OmitEmptyTest1: "not empty",
			OmitEmptyTest2: "",
			OmitEmptyTest3: "",
			OmitEmptyTest4: "not empty",
		})
	if err != nil {
		t.Fatal(err)
	}
	if err = enc.Encode(res); err != nil {
		t.Fatal(err)
	}
	f := &fluxast.File{}
	err = f.UnmarshalJSON(buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	formatted := fluxast.Format(f)
	_, err = flux.Parse(formatted)
	if err != nil {
		t.Fatal(err)
	}
	expected := `fred = {cow: {"9": true}, B: 1776}
complex = {"hello": {}, "okra": /[a-z]+/, "mapmapmap": {"map0": {"map1": {"map2": 0}}}}
hello = "x"
hellofriend = 3
uinsignedIntTest = 7
nowabool = false
Public = {"perl harbor day": 1941-12-07T07:48:00-10:30}
OmitEmptyTest1 = "not empty"
omit_empty_test_snake2 = "not empty"`

	// we do this because maps can have quazi random order
	splitFomattedLines := strings.Split(formatted, "\n")
	sort.Strings(splitFomattedLines)
	splitExpectedLines := strings.Split(expected, "\n")
	sort.Strings(splitExpectedLines)
	if strings.Join(splitFomattedLines, "\n") != strings.Join(splitExpectedLines, "\n") {
		t.Errorf("expected: \n%s\n but got \n%s\n", expected, formatted)
	}

}
