
package encoding

import (
	"math"
)

func EncodeUint16(buf []byte, v uint16)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	return buf
}

func EncodeUint32(buf []byte, v uint32)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	buf[2] = (byte)(v >> 16 & 0xff)
	buf[3] = (byte)(v >> 24 & 0xff)
	return buf
}

func EncodeUint64(buf []byte, v uint64)([]byte){
	buf[0] = (byte)(v & 0xff)
	buf[1] = (byte)(v >> 8 & 0xff)
	buf[2] = (byte)(v >> 16 & 0xff)
	buf[3] = (byte)(v >> 24 & 0xff)
	buf[4] = (byte)(v >> 32 & 0xff)
	buf[5] = (byte)(v >> 40 & 0xff)
	buf[6] = (byte)(v >> 48 & 0xff)
	buf[7] = (byte)(v >> 56 & 0xff)
	return buf
}

func EncodeFloat32(buf []byte, v float32)([]byte){
	return EncodeUint32(buf, math.Float32bits(v))
}

func EncodeFloat64(buf []byte, v float64)([]byte){
	return EncodeUint64(buf, math.Float64bits(v))
}


func DecodeUint16(buf []byte)(uint16){
	return (uint16)(buf[0]) | (uint16)(buf[1]) << 8
}

func DecodeUint32(buf []byte)(uint32){
	return (uint32)(buf[0]) | (uint32)(buf[1]) << 8 |
		(uint32)(buf[2]) << 16 | (uint32)(buf[3]) << 24
}

func DecodeUint64(buf []byte)(uint64){
	return (uint64)(buf[0]) | (uint64)(buf[1]) << 8 |
		(uint64)(buf[2]) << 16 | (uint64)(buf[3]) << 24 |
		(uint64)(buf[4]) << 32 | (uint64)(buf[5]) << 40 |
		(uint64)(buf[6]) << 48 | (uint64)(buf[7]) << 56
}

func DecodeFloat32(buf []byte)(float32){
	return math.Float32frombits(DecodeUint32(buf))
}

func DecodeFloat64(buf []byte)(float64){
	return math.Float64frombits(DecodeUint64(buf))
}


func EncodeUint16s(buf []byte, v []uint16)([]byte){
	for i, n := range v {
		EncodeUint16(buf[i * 2:], n)
	}
	return buf
}

func EncodeUint32s(buf []byte, v []uint32)([]byte){
	for i, n := range v {
		EncodeUint32(buf[i * 4:], n)
	}
	return buf
}

func EncodeUint64s(buf []byte, v []uint64)([]byte){
	for i, n := range v {
		EncodeUint64(buf[i * 8:], n)
	}
	return buf
}

func EncodeFloat32s(buf []byte, v []float32)([]byte){
	for i, n := range v {
		EncodeFloat32(buf[i * 4:], n)
	}
	return buf
}

func EncodeFloat64s(buf []byte, v []float64)([]byte){
	for i, n := range v {
		EncodeFloat64(buf[i * 8:], n)
	}
	return buf
}

func DecodeUint16s(buf []byte)(v []uint16){
	v = make([]uint16, len(buf) / 2)
	for i, _ := range v {
		v[i] = DecodeUint16(buf[i * 2:])
	}
	return
}

func DecodeUint32s(buf []byte)(v []uint32){
	v = make([]uint32, len(buf) / 4)
	for i, _ := range v {
		v[i] = DecodeUint32(buf[i * 4:])
	}
	return
}

func DecodeUint64s(buf []byte)(v []uint64){
	v = make([]uint64, len(buf) / 8)
	for i, _ := range v {
		v[i] = DecodeUint64(buf[i * 8:])
	}
	return
}

func DecodeFloat32s(buf []byte)(v []float32){
	v = make([]float32, len(buf) / 4)
	for i, _ := range v {
		v[i] = DecodeFloat32(buf[i * 4:])
	}
	return
}

func DecodeFloat64s(buf []byte)(v []float64){
	v = make([]float64, len(buf) / 8)
	for i, _ := range v {
		v[i] = DecodeFloat64(buf[i * 8:])
	}
	return
}

