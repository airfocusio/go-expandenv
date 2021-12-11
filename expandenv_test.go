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
	values := map[string]string{
		"ENV_A":          "a",
		"ENV_B":          "b",
		"ENV_42":         "42",
		"ENV_42_5":       "42.5",
		"ENV_YES":        "yes",
		"ENV_MULTI_LINE": "line1\nline2",
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
			input:  "\\${ENV_A}",
			output: "${ENV_A}",
			label:  "variabled-escacped-string",
		},
		{
			input:  "prefix \\${ENV_A} suffix",
			output: "prefix ${ENV_A} suffix",
			label:  "variabled-escacped-string-2",
		},
		{
			input:  map[string]interface{}{"single": "some ${ENV_A}", "multi": "some ${ENV_MULTI_LINE}"},
			output: map[string]interface{}{"single": "some a", "multi": "some line1\nline2"},
			label:  "variabled-nested",
		},
		{
			input:  "${ENV_42}",
			output: "42",
			label:  "variabled-format",
		},
		{
			input:  "${ENV_42:number}",
			output: 42,
			label:  "variabled-format-2",
		},
		{
			input:  "${ENV_42_5:number}",
			output: 42.5,
			label:  "variabled-format-3",
		},
		{
			input:  "${ENV_YES:boolean}",
			output: true,
			label:  "variabled-format-4",
		},
		{
			input:  "${ENV_42:boolean}",
			output: "${ENV_42:boolean}",
			label:  "variabled-format-2",
			error:  fmt.Errorf("42 is not a valid boolean"),
		},
		{
			input:  "foo: some ${ENV_A} ${ENV_UNKNOWN}",
			output: "foo: some a ${ENV_UNKNOWN}",
			label:  "variabled-unknown",
			error:  fmt.Errorf("environment variable ENV_UNKNOWN is missing"),
		},
		{
			input:  "foo: some ${ENV_A} |${ENV_B:-fallback}|",
			output: "foo: some a |b|",
			label:  "variabled-fallback",
		},
		{
			input:  "foo: some ${ENV_A} |${ENV_UNKNOWN:-fallback}|",
			output: "foo: some a |fallback|",
			label:  "variabled-fallback-1",
		},
		{
			input:  "foo: some ${ENV_A} |${ENV_UNKNOWN:-}|",
			output: "foo: some a ||",
			label:  "variabled-fallback-2",
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

func TestExpandWithYaml(t *testing.T) {
	values := map[string]string{
		"ENV_A":          "a",
		"ENV_MULTI_LINE": "line1\nline2",
	}

	yamlBytes := []byte(`
a: ${ENV_A}
b: prefix ${ENV_A} suffix
c:
    - ${ENV_A}
    - ${ENV_A}
d: ${ENV_MULTI_LINE}
e: ${ENV_UNKNOWN}
f: \${ENV_ESCAPED}
g: \\${ENV_ESCAPED}
`)
	var yamlRaw interface{}
	err := yaml.Unmarshal(yamlBytes, &yamlRaw)
	assert.NoError(t, err)
	yamlRaw, err = Expand(yamlRaw, values)
	assert.EqualError(t, err, "environment variable ENV_UNKNOWN is missing")
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
e: ${ENV_UNKNOWN}
f: ${ENV_ESCAPED}
g: \${ENV_ESCAPED}
`, string(yamlBytes))
}
