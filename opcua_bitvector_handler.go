package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/gopcua/opcua/ua"
	"github.com/prometheus/client_golang/prometheus"
)

// OpcuaBitVectorHandler extracts a single bit from an UPCUA Variant value
// and sets a prometheus guage to the coresponding value: 0.0 or 1.0.
// The bit indexing starts at zero, which represents the least significant bit.
// So, for the 32-bit hex value 0xF0F10F0F, if we're asked for bit 16,
// we want the bit shown below in parentheses:
// 11110000 1111000(1) 00001111 00001111
type OpcuaBitVectorHandler struct {
	gauge      prometheus.Gauge
	extractBit int // identifies the bit to extract. little endian bit & byte order.
	debug      bool
}

// Handle computes the float value and emit it as a prometheus metric.
func (h OpcuaBitVectorHandler) Handle(v ua.Variant) error {
	floatVal, err := h.FloatValue(v)
	if err != nil {
		return err
	}
	if h.debug {
		log.Printf("Extracted bit number %d: value=%d", h.extractBit, int(floatVal))
	}
	h.gauge.Set(floatVal)
	return nil
}

// FloatValue returns the value of the requested bit within the Variant value
func (h OpcuaBitVectorHandler) FloatValue(v ua.Variant) (float64, error) {
	var bytes []byte
	var err error

	bytes, err = variantToByteArray(v)
	if err != nil {
		return 0.0, err
	}

	bitValue, extractErr := extractBit(bytes, h.extractBit)
	if extractErr != nil {
		return 0.0, extractErr
	}
	return float64(bitValue), nil
}

/**
* Convert a fixed-length variant value to a little-endian byte array
**/
func variantToByteArray(v ua.Variant) ([]byte, error) {
	value := v.Value()
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.LittleEndian, value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

/**
* Extract a single bit from an integer value of unknown format, return a zero or one.
* Input bytes are expected to by in little-endian order.
* Returns an error if the value cannot be interpreted as some sort of integer.
**/
func extractBit(bytes []byte, bit int) (byte, error) {
	if bit < 0 {
		return 0, fmt.Errorf("Bit number must be positive. Got %d", bit)
	}

	// decompose bit number into a byte index and a bit index within that byte
	byteIdx := bit / 8
	bitIdx := bit % 8
	if byteIdx > len(bytes)-1 {
		return 0, fmt.Errorf("Bit %d out of range for %d-byte value", bit, len(bytes))
	}
	bite := bytes[byteIdx]
	bitValue := (bite & (0x01 << bitIdx)) >> bitIdx
	return bitValue, nil
}
