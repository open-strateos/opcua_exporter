package main

import (
	"testing"
	"time"

	"github.com/gopcua/opcua/ua"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func getTestExtractHandler(extractBit int) OpcuaBitVectorHandler {
	g := prom.NewGauge(prom.GaugeOpts{Name: "bar"})
	return OpcuaBitVectorHandler{g, extractBit, false}
}

func TestHandleBitVector(t *testing.T) {
	type testCase struct {
		value interface{}
		bit   int
		want  float64
	}

	testCases := []testCase{
		{uint8(0x01), 0, 1},
		{uint8(0x04), 0, 0},
		{uint8(0x04), 2, 1},
		{uint16(128), 7, 1},
		{uint16(128), 6, 0},
		{uint16(0x0100), 8, 1.0},
		{uint32(128), 7, 1},
		{uint64(128), 7, 1},
		{uint32(65536), 16, 1},
		{uint32(65537), 16, 1},
		{uint64(65538), 16, 1},
		{uint32(0x00100098), 16, 0},
		{uint32(0x00100098), 20, 1},
		{uint32(0x01000000), 24, 1},
		{int8(0x01), 0, 1},
		{int8(0x04), 0, 0},
		{int8(0x04), 2, 1},
		{int16(128), 7, 1},
		{int16(128), 6, 0},
		{int16(0x0100), 8, 1.0},
		{int32(128), 7, 1},
		{int64(128), 7, 1},
		{int32(65536), 16, 1},
		{int32(65537), 16, 1},
		{int64(65538), 16, 1},
		{int32(0x00100098), 16, 0},
		{int32(0x00100098), 20, 1},
		{int32(0x01000000), 24, 1},
		{int64(0x0100000000000000), 56, 1},
	}

	for _, tc := range testCases {
		handler := getTestExtractHandler(tc.bit)
		variant, variantErr := ua.NewVariant(tc.value)
		assert.Nil(t, variantErr)
		floatVal, err := handler.FloatValue(*variant)
		assert.Nil(t, err)
		assert.Equal(t, tc.want, floatVal)

	}

	// Various non-integer objects shold error out
	type errorCase struct {
		value      interface{}
		extractBit int
	}
	errorCases := []errorCase{
		{"Not a number", 4},
		{time.Now(), 4}, // valid Variant value, but not valid for binary.Write()
		{int16(78), -2}, // negative bit index
	}
	for _, tc := range errorCases {
		handler := getTestExtractHandler(tc.extractBit)
		variant, variantErr := ua.NewVariant(tc.value)
		assert.Nil(t, variantErr)
		floatVal, err := handler.FloatValue(*variant)
		assert.Error(t, err)
		assert.Equal(t, 0.0, floatVal)

	}
}

func TestVariantToByteArray(t *testing.T) {
	type testCase struct {
		value interface{}
		want  []byte
	}

	testCases := []testCase{
		{int8(0x38), []byte{0x38}},
		{int16(0x0000), []byte{0x00, 0x00}},
		{int16(0x0102), []byte{0x02, 0x01}},
		{int16(0x0001), []byte{0x01, 0x00}},
		{int16(-1), []byte{0xff, 0xff}},
		{int16(-4), []byte{0xfc, 0xff}},
		{int32(0x0001), []byte{0x01, 0x00, 0x00, 0x00}},
		{int32(0x00010001), []byte{0x01, 0x00, 0x01, 0x00}},
		{int64(0x0100000000000000), []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
	}

	for _, tc := range testCases {
		variant, _ := ua.NewVariant(tc.value)
		bytes, err := variantToByteArray(*variant)
		assert.Nil(t, err)
		assert.Equal(t, tc.want, bytes)
	}
}
