package utils

import "unsafe"

func Float32SliceToBytes(slice []float32) []byte {
	return (*[1 << 31]byte)(unsafe.Pointer(&slice[0]))[: len(slice)*4 : len(slice)*4]
}

func BytesToFloat32Slice(bytes []byte) []float32 {
	return (*[1 << 31]float32)(unsafe.Pointer(&bytes[0]))[: len(bytes)/4 : len(bytes)/4]
}
