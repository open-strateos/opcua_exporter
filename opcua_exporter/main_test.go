package main

import (
	"testing"

	"github.com/gopcua/opcua/monitor"
	"github.com/gopcua/opcua/ua"
	"github.com/stretchr/testify/assert"
)

/**
* mockHandler implements the MsgHandler interface
 */
type mockHandler struct {
	called bool // set if the Handle() func has been called
}

func (th *mockHandler) Handle(val ua.Variant) error {
	th.called = true
	return nil
}

func (th *mockHandler) FloatValue(val ua.Variant) (float64, error) {
	return 0.0, nil
}

func makeTestMessage(nodeID *ua.NodeID) monitor.DataChangeMessage {
	return monitor.DataChangeMessage{
		NodeID: nodeID,
		DataValue: &ua.DataValue{
			Value: ua.MustVariant(0.0),
		},
	}
}

// Excercise the handleMessage() function
// Ensure that it dispatches messages to handlers as expected
func TestHandleMessage(t *testing.T) {
	nodeID1 := ua.NewStringNodeID(1, "foo")
	nodeID2 := ua.NewStringNodeID(1, "bar")
	nodeName1 := nodeID1.String()
	nodeName2 := nodeID2.String()
	handlerMap := make(HandlerMap)
	for i := 0; i < 3; i++ {
		mapRecord := handlerMapRecord{
			config:  NodeConfig{NodeName: nodeName1, MetricName: "whatever"},
			handler: &mockHandler{},
		}
		handlerMap[nodeName1] = append(handlerMap[nodeName1], mapRecord)
	}

	mapRecord2 := handlerMapRecord{
		config:  NodeConfig{NodeName: nodeName2, MetricName: "whatever"},
		handler: &mockHandler{},
	}
	handlerMap[nodeName2] = append(handlerMap[nodeName2], mapRecord2)

	assert.Equal(t, len(handlerMap[nodeName1]), 3)
	assert.Equal(t, len(handlerMap[nodeName2]), 1)

	// Handle a fake message addressed to nodeID1
	msg := makeTestMessage(nodeID1)
	handleMessage(&msg, handlerMap)

	// All three nodeName1 handlers should have been called
	for _, record := range handlerMap[nodeName1] {
		handler := record.handler.(*mockHandler)
		assert.True(t, handler.called)
	}

	// But not the nodeName2 handler
	handler := handlerMap[nodeName2][0].handler.(*mockHandler)
	assert.False(t, handler.called)

}

// Exercise the createMetrics() function
// Ensure that it creats the right sort of HandlerMap
func TestCreateMetrics(t *testing.T) {
	nodeconfigs := []NodeConfig{
		{
			NodeName:   "foo",
			MetricName: "foo_level_blorbs",
		},
		{
			NodeName:   "bar",
			MetricName: "bar_level_blorbs",
		},
		{
			NodeName:   "foo",
			MetricName: "foo_rate_blarbs",
		},
	}

	handlerMap := createMetrics(&nodeconfigs)
	assert.Equal(t, len(handlerMap), 2)
	assert.Equal(t, len(handlerMap["foo"]), 2)
	assert.Equal(t, len(handlerMap["bar"]), 1)
}
