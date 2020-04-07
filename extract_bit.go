package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

/**
* Extract a single bit from an integer value of unknown format, return a zero or one.
* Returns an error if the value cannot be interpreted as some sort of integer.
**/
func extractBit(input interface{}, bit uint8) (uint8, error) {
	bytes, typeError := uintToByteString(input)
	if typeError != nil {
		return 0, typeError
	}

	// decompose bit number into a byte index and a bit index within that byte
	byteIdx := bit / 8
	bitIdx := bit % 8
	if byteIdx > uint8(len(bytes)-1) {
		return 0, fmt.Errorf("Bit %d out of range for type %s ", bit, reflect.TypeOf(input))
	}
	bite := bytes[byteIdx]
	result := (bite & (0x01 << bitIdx)) >> bitIdx
	return result, nil
}

/**
* Convert an unsigned integer to a little-endian byte array
**/
func uintToByteString(unknown interface{}) ([]byte, error) {
	var buf []byte
	switch t := unknown.(type) {
	case uint8:
		buf = []byte{byte(unknown.(uint8))}
	case uint16:
		size := 16
		buf = make([]byte, size/8)
		binary.LittleEndian.PutUint16(buf, uint16(unknown.(uint16)))
	case uint32:
		size := 32
		buf = make([]byte, size/8)
		binary.LittleEndian.PutUint32(buf, uint32(unknown.(uint32)))
	case uint64:
		size := 64
		buf = make([]byte, size/8)
		binary.LittleEndian.PutUint64(buf, uint64(unknown.(uint64)))
	default:
		return nil, fmt.Errorf("Type was %s, butintToByteString only handles unsigned integer types", reflect.TypeOf(t))
	}
	return buf, nil
}
