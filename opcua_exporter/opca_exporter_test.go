package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testNodes = []Node{
	Node{
		NodeName:   "foo",
		MetricName: "bar",
	},
}

func TestReadNodeFile(t *testing.T) {
	//testNodes := strings.NewReader(`[{"nodeName": "foo", "metricName": "bar"} ]`)
	data, _ := json.Marshal(testNodes)
	results, err := parseConfigJSON(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, len(testNodes), len(results))
	assert.IsType(t, Node{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)

	results, err = parseConfigJSON(strings.NewReader("foooob not valid json here"))
	assert.Error(t, err)
	assert.Empty(t, results)
}

func TestB64Config(t *testing.T) {
	data, _ := json.Marshal(testNodes)
	encodedData := base64.StdEncoding.EncodeToString(data)
	results, err := readConfigBase64(&encodedData)
	assert.NoError(t, err)
	assert.Equal(t, len(testNodes), len(results))
	assert.IsType(t, Node{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)
}

type floatTest struct {
	input  interface{}
	output float64
}

func TestCoerceToFloat(t *testing.T) {
	testCases := []floatTest{
		floatTest{2, 2.0},
		floatTest{int64(25), 25.0},
		floatTest{int32(33), 33.0},
		floatTest{true, 1.0},
		floatTest{false, 0.0},
		floatTest{float32(8.8), float64(float32(8.8))}, // float32 --> float64 actually introduces rounding errors on the order of 1e-7
	}
	for _, testCase := range testCases {
		result, err := coerceToFloat64(testCase.input)
		assert.Nil(t, err)
		assert.Equal(t, testCase.output, result)
	}

	_, err := coerceToFloat64("not a number")
	assert.Error(t, err)

}
