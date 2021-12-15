package expandenv

import (
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestExpand(t *testing.T) {
	values := func(key string) (*string, error) {
		switch key {
		case "FN_A":
			result := "a"
			return &result, nil
		case "FN_B":
			result := "b"
			return &result, nil
		case "FN_42":
			result := "42"
			return &result, nil
		case "FN_42_5":
			result := "42.5"
			return &result, nil
		case "FN_YES":
			result := "yes"
			return &result, nil
		case "FN_MULTI_LINE":
			result := "line1\nline2"
			return &result, nil
		case "FN_IGNORE":
			return nil, nil
		default:
			return nil, fmt.Errorf("unknown")
		}
	}

	testCases := []struct {
		input  interface{}
		output interface{}
		label  string
		error  error
	}{
		{
			input:  nil,
			output: nil,
			label:  "static-empty",
		},
		{
			input:  "string",
			output: "string",
			label:  "static-string",
		},
		{
			input:  1,
			output: 1,
			label:  "static-int",
		},
		{
			input:  math.MaxInt64,
			output: math.MaxInt64,
			label:  "static-int-large",
		},
		{
			input:  1.2,
			output: 1.2,
			label:  "static-double",
		},
		{
			input:  []interface{}{"a", 1},
			output: []interface{}{"a", 1},
			label:  "static-array",
		},
		{
			input:  map[string]interface{}{"a": "b", "c": 2},
			output: map[string]interface{}{"a": "b", "c": 2},
			label:  "static-map",
		},
		{
			input:  "${FN_A}",
			output: "a",
			label:  "variabled-string",
		},
		{
			input:  "prefix ${FN_B} suffix",
			output: "prefix b suffix",
			label:  "variabled-string-2",
		},
		{
			input:  "\\${FN_A}",
			output: "${FN_A}",
			label:  "variabled-escacped-string",
		},
		{
			input:  "prefix \\${FN_A} suffix",
			output: "prefix ${FN_A} suffix",
			label:  "variabled-escacped-string-2",
		},
		{
			input:  map[string]interface{}{"single": "some ${FN_A}", "multi": "some ${FN_MULTI_LINE}"},
			output: map[string]interface{}{"single": "some a", "multi": "some line1\nline2"},
			label:  "variabled-nested",
		},
		{
			input:  "${FN_42}",
			output: "42",
			label:  "variabled-format",
		},
		{
			input:  "${FN_42:number}",
			output: 42,
			label:  "variabled-format-2",
		},
		{
			input:  "${FN_42_5:number}",
			output: 42.5,
			label:  "variabled-format-3",
		},
		{
			input:  "${FN_YES:boolean}",
			output: true,
			label:  "variabled-format-4",
		},
		{
			input:  "${FN_42:boolean}",
			output: "${FN_42:boolean}",
			label:  "variabled-format-2",
			error:  fmt.Errorf("42 is not a valid boolean"),
		},
		{
			input:  "foo: some ${FN_A} ${FN_UNKNOWN}",
			output: "foo: some a ${FN_UNKNOWN}",
			label:  "variabled-unknown",
			error:  fmt.Errorf("unknown"),
		},
		{
			input:  "foo: some ${FN_A} |${FN_B:-fallback}|",
			output: "foo: some a |b|",
			label:  "variabled-fallback",
		},
		{
			input:  "foo: some ${FN_A} |${FN_UNKNOWN:-fallback}|",
			output: "foo: some a |fallback|",
			label:  "variabled-fallback-1",
		},
		{
			input:  "foo: some ${FN_A} |${FN_UNKNOWN:-}|",
			output: "foo: some a ||",
			label:  "variabled-fallback-2",
		},
		{
			input:  "foo: ${FN_IGNORE}",
			output: "foo: ${FN_IGNORE}",
			label:  "variabled-ignored",
		},
	}

	for _, testCase := range testCases {
		output, err := Expand(testCase.input, values)
		if testCase.error == nil {
			assert.NoError(t, err, testCase.label)
		} else {
			assert.EqualError(t, err, testCase.error.Error(), testCase.label)
		}
		assert.Equal(t, testCase.output, output, testCase.label)
	}
}

func TestExpandMap(t *testing.T) {
	values := map[string]string{
		"MAP_A": "a",
		"MAP_B": "b",
	}

	testCases := []struct {
		input  interface{}
		output interface{}
		label  string
		error  error
	}{
		{
			input:  "${MAP_A}",
			output: "a",
			label:  "variabled-string",
		},
		{
			input:  "prefix ${MAP_B} suffix",
			output: "prefix b suffix",
			label:  "variabled-string-2",
		},
		{
			input:  "foo: some ${MAP_A} ${MAP_UNKNOWN}",
			output: "foo: some a ${MAP_UNKNOWN}",
			label:  "variabled-unknown",
			error:  fmt.Errorf("variable MAP_UNKNOWN is missing"),
		},
	}

	for _, testCase := range testCases {
		output, err := ExpandMap(testCase.input, values)
		if testCase.error == nil {
			assert.NoError(t, err, testCase.label)
		} else {
			assert.EqualError(t, err, testCase.error.Error(), testCase.label)
		}
		assert.Equal(t, testCase.output, output, testCase.label)
	}
}

func TestExpandEnv(t *testing.T) {
	os.Setenv("ENV_A", "a")
	os.Setenv("ENV_B", "b")

	testCases := []struct {
		input  interface{}
		output interface{}
		label  string
		error  error
	}{
		{
			input:  "${ENV_A}",
			output: "a",
			label:  "variabled-string",
		},
		{
			input:  "prefix ${ENV_B} suffix",
			output: "prefix b suffix",
			label:  "variabled-string-2",
		},
		{
			input:  "foo: some ${ENV_A} ${ENV_UNKNOWN}",
			output: "foo: some a ${ENV_UNKNOWN}",
			label:  "variabled-unknown",
			error:  fmt.Errorf("environment variable ENV_UNKNOWN is missing"),
		},
	}

	for _, testCase := range testCases {
		output, err := ExpandEnv(testCase.input)
		if testCase.error == nil {
			assert.NoError(t, err, testCase.label)
		} else {
			assert.EqualError(t, err, testCase.error.Error(), testCase.label)
		}
		assert.Equal(t, testCase.output, output, testCase.label)
	}
}

func TestExpandMapWithYaml(t *testing.T) {
	values := map[string]string{
		"MAP_A":          "a",
		"MAP_MULTI_LINE": "line1\nline2",
	}

	yamlBytes := []byte(`
a: ${MAP_A}
b: prefix ${MAP_A} suffix
c:
    - ${MAP_A}
    - ${MAP_A}
d: ${MAP_MULTI_LINE}
e: ${MAP_UNKNOWN}
f: \${MAP_ESCAPED}
g: \\${MAP_ESCAPED}
`)
	var yamlRaw interface{}
	err := yaml.Unmarshal(yamlBytes, &yamlRaw)
	assert.NoError(t, err)
	yamlRaw, err = ExpandMap(yamlRaw, values)
	assert.EqualError(t, err, "variable MAP_UNKNOWN is missing")
	yamlBytes, err = yaml.Marshal(yamlRaw)
	assert.NoError(t, err)
	assert.Equal(t, `a: a
b: prefix a suffix
c:
    - a
    - a
d: |-
    line1
    line2
e: ${MAP_UNKNOWN}
f: ${MAP_ESCAPED}
g: \${MAP_ESCAPED}
`, string(yamlBytes))
}
