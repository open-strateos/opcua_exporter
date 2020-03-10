package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadNodeFile(t *testing.T) {
	testNodes := strings.NewReader(`[{"nodeName": "foo", "metricName": "bar"} ]`)
	results, err := parseNodeJSONFile(testNodes)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(results))
	assert.IsType(t, Node{}, results[0])
	assert.Equal(t, "foo", results[0].NodeName)
	assert.Equal(t, "bar", results[0].MetricName)

	results, err = parseNodeJSONFile(strings.NewReader("foooob not valid json here"))
	assert.Error(t, err)
	assert.Empty(t, results)
}
