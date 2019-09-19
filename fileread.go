package main

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"unicode/utf16"
)

func readInt64(r io.Reader) uint64 {
	bytes, _ := readNextBytes(r, 8)
	return binary.LittleEndian.Uint64(bytes)
}

func readInt32(r io.Reader) uint32 {
	bytes, _ := readNextBytes(r, 4)
	return binary.LittleEndian.Uint32(bytes)
}

func readInt16(r io.Reader) uint16 {
	bytes, _ := readNextBytes(r, 2)
	return binary.LittleEndian.Uint16(bytes)
}

func readInt8(r io.Reader) (uint8, error) {
	bytes, err := readNextBytes(r, 1)
	return bytes[0], err

}

func readStruct(r io.Reader, mystruct interface{}, size uint32) {

	data, _ := readNextBytes(r, size)

	buffer := bytes.NewBuffer(data)
	err := binary.Read(buffer, binary.LittleEndian, mystruct)
	if err != nil {
		log.Fatal("binary.Read failed", err)
	}
}

func readNextBytes(buffer io.Reader, number uint32) ([]byte, error) {
	bytes := make([]byte, number)

	_, err := buffer.Read(bytes)
	if err != nil {
		if err == io.EOF {
			return bytes, err
		}
		log.Fatal("Error reading file:", err)
	}
	return bytes, nil
}

func unicodeToString(data []byte) string {
	utf := make([]uint16, len(data)/2)

	for i := 0; i < len(data); i += 2 {
		utf[i/2] = binary.LittleEndian.Uint16(data[i:])
	}

	return string(utf16.Decode(utf))
}
