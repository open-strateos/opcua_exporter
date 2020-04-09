package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testNodes = []NodeConfig{
	NodeConfig{
		NodeName:   "foo",
		MetricName: "bar",
	},
	NodeConfig{
		NodeName:   "baz",
		MetricName: "bak",
		ExtractBit: 4,
	},
}

func TestReadNodeFile(t *testing.T) {
	//testNodes := strings.NewReader(`[{"nodeName": "foo", "metricName": "bar"} ]`)
	data, _ := json.Marshal(testNodes)
	results, err := parseConfigJSON(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, len(testNodes), len(results))
	assert.IsType(t, NodeConfig{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)

	assert.Nil(t, results[0].ExtractBit)
	assert.Equal(t, 4.0, results[1].ExtractBit) // float64, because json

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
	assert.IsType(t, NodeConfig{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)
}

type floatTest struct {
	input  interface{}
	output float64
}

func TestExtractBit(t *testing.T) {
	type extractTest struct {
		value    interface{}
		bit      int
		expected uint8
	}

	// Should be able to process any of the 3 unsigned integer types
	testCases := []extractTest{
		extractTest{uint8(4), 0, 0},
		extractTest{uint8(4), 2, 1},
		extractTest{uint16(128), 7, 1},
		extractTest{uint16(128), 6, 0},
		extractTest{uint32(128), 7, 1},
		extractTest{uint64(128), 7, 1},
		extractTest{uint32(65536), 16, 1},
		extractTest{uint32(65537), 16, 1},
		extractTest{uint64(65538), 16, 1},
		extractTest{uint32(0x00100098), 16, 0},
		extractTest{uint32(0x00100098), 20, 1},
		extractTest{uint32(0x01000000), 24, 1},
	}

	for _, testCase := range testCases {
		result, err := extractBit(testCase.value, testCase.bit)
		assert.Nil(t, err)
		assert.Equal(t, testCase.expected, result)
	}

	// Things that don't work here.
	errorCases := []extractTest{
		extractTest{0x11, -3, 0},          // bit argument is negative
		extractTest{uint16(32768), 22, 0}, // bit out of range
		extractTest{2.2, 7, 0},            // not an integer
		extractTest{int16(3), 2, 0},       // signed integer
		extractTest{32, 2, 0},             // signed integer
		extractTest{"foo", 3, 0},          // string
	}

	for _, errorCase := range errorCases {
		result, err := extractBit(errorCase.value, errorCase.bit)
		assert.Equal(t, uint8(0), result)
		assert.Error(t, err)
	}
}
