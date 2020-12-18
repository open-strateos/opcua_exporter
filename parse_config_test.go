package main

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

var testNodes = []NodeConfig{
	{
		NodeName:   "foo",
		MetricName: "bar",
	},
	{
		NodeName:   "baz",
		MetricName: "bak",
		ExtractBit: 4,
	},
}

func TestReadNodeFile(t *testing.T) {
	data, _ := yaml.Marshal(testNodes)
	results, err := parseConfigYAML(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, len(testNodes), len(results))
	assert.IsType(t, NodeConfig{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)

	assert.Nil(t, results[0].ExtractBit)
	assert.Equal(t, 4, results[1].ExtractBit)

	results, err = parseConfigYAML(strings.NewReader("foooob not valid json here"))
	assert.Error(t, err)
	assert.Empty(t, results)
}

func TestB64Config(t *testing.T) {
	data, _ := yaml.Marshal(testNodes)
	encodedData := base64.StdEncoding.EncodeToString(data)
	results, err := readConfigBase64(&encodedData)
	assert.NoError(t, err)
	assert.Equal(t, len(testNodes), len(results))
	assert.IsType(t, NodeConfig{}, results[0])
	assert.Equal(t, testNodes[0].NodeName, results[0].NodeName)
	assert.Equal(t, testNodes[0].MetricName, results[0].MetricName)
}

func TestJsonConfig(t *testing.T) {
	// Luckily JSON is also YAML
	json := `[{"metricName": "foo", "nodeName": "whatever", "extractBit": 1}]`
	results, err := parseConfigYAML(strings.NewReader((json)))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.Equal(t, "foo", results[0].MetricName)
	assert.Equal(t, "whatever", results[0].NodeName)
	assert.Equal(t, 1, results[0].ExtractBit)
}
