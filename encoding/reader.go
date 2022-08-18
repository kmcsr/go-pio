
package encoding

import (
	"io"
)

type (
	Reader interface{
		io.Closer
		io.Reader
		ReadBool()(v bool, err error)
		ReadByte()(v byte, err error)
		ReadUint16()(v uint16, err error)
		ReadUint32()(v uint32, err error)
		ReadUint64()(v uint64, err error)
		ReadFloat32()(v float32, err error)
		ReadFloat64()(v float64, err error)
		ReadString()(v string, err error)
		ReadBools()(v []bool, err error)
		ReadBytes()(v []byte, err error)
		ReadUint16s()(v []uint16, err error)
		ReadUint32s()(v []uint32, err error)
		ReadUint64s()(v []uint64, err error)
	}
	reader struct{
		io.Reader
	}
)

var _ Reader = (*reader)(nil)

func WrapReader(old io.Reader)(Reader){
	if r, ok := old.(Reader); ok {
		return r
	}
	return &reader{
		Reader: old,
	}
}

func (r *reader)Close()(error){
	if c, ok := r.Reader.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (r *reader)ReadBool()(v bool, err error){
	var v0 byte
	v0, err = r.ReadByte()
	if err != nil {
		return
	}
	v = v0 != 0
	return
}

func (r *reader)ReadByte()(v byte, err error){
	var buf [1]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = buf[0]
	return
}

func (r *reader)ReadUint16()(v uint16, err error){
	var buf [2]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = DecodeUint16(buf[:])
	return
}

func (r *reader)ReadUint32()(v uint32, err error){
	var buf [4]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = DecodeUint32(buf[:])
	return
}

func (r *reader)ReadUint64()(v uint64, err error){
	var buf [8]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = DecodeUint64(buf[:])
	return
}

func (r *reader)ReadFloat32()(v float32, err error){
	var buf [4]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = DecodeFloat32(buf[:])
	return
}

func (r *reader)ReadFloat64()(v float64, err error){
	var buf [8]byte
	_, err = io.ReadFull(r.Reader, buf[:])
	if err != nil {
		return
	}
	v = DecodeFloat64(buf[:])
	return
}

func (r *reader)ReadString()(v string, err error){
	var v0 []byte
	v0, err = r.ReadBytes()
	if err != nil {
		return
	}
	v = (string)(v0)
	return
}

func (r *reader)ReadBools()(v []bool, err error){
	var v0 []byte
	v0, err = r.ReadBytes()
	if err != nil {
		return
	}
	v = make([]bool, len(v0))
	for i, o := range v0 {
		v[i] = o != 0
	}
	return
}

func (r *reader)ReadBytes()(v []byte, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	v = make([]byte, l)
	_, err = io.ReadFull(r.Reader, v)
	return
}

func (r *reader)ReadUint16s()(v []uint16, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	buf := make([]byte, l * 2)
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return
	}
	v = DecodeUint16s(buf)
	return
}

func (r *reader)ReadUint32s()(v []uint32, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	buf := make([]byte, l * 4)
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return
	}
	v = DecodeUint32s(buf)
	return
}

func (r *reader)ReadUint64s()(v []uint64, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	buf := make([]byte, l * 8)
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return
	}
	v = DecodeUint64s(buf)
	return
}

func (r *reader)ReadFloat32s()(v []float32, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	buf := make([]byte, l * 4)
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return
	}
	v = DecodeFloat32s(buf)
	return
}

func (r *reader)ReadFloat64s()(v []float64, err error){
	var l uint32
	l, err = r.ReadUint32()
	if err != nil {
		return
	}
	buf := make([]byte, l * 8)
	_, err = io.ReadFull(r.Reader, buf)
	if err != nil {
		return
	}
	v = DecodeFloat64s(buf)
	return
}

