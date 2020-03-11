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
