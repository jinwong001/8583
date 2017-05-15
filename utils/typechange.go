package utils

import (
	"bytes"
	"encoding/binary"
	"strings"
	"encoding/hex"
)

func Byte2Int(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

func Int2Byte(i int) []byte {
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, i)
	return bytesBuffer.Bytes()
}

func EncodeToString(src []byte) string {
	return strings.ToUpper(hex.EncodeToString(src))
}