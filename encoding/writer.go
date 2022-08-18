
package encoding

import (
	"io"
)

type (
	Writer interface{
		io.Closer
		io.Writer
		WriteBool(v bool)(error)
		WriteByte(v byte)(error)
		WriteUint16(v uint16)(error)
		WriteUint32(v uint32)(error)
		WriteUint64(v uint64)(error)
		WriteFloat32(v float32)(error)
		WriteFloat64(v float64)(error)
		WriteString(v string)(error)
		WriteBools(v []bool)(error)
		WriteBytes(v []byte)(error)
		WriteUint16s(v []uint16)(error)
		WriteUint32s(v []uint32)(error)
		WriteUint64s(v []uint64)(error)
		WriteFloat32s(v []float32)(error)
		WriteFloat64s(v []float64)(error)
	}
	writer struct{
		io.Writer
	}
)

var _ Writer = (*writer)(nil)

func WrapWriter(old io.Writer)(Writer){
	if w, ok := old.(Writer); ok {
		return w
	}
	return &writer{
		Writer: old,
	}
}

func Pipe()(r Reader, w Writer){
	ir, iw := io.Pipe()
	r, w = WrapReader(ir), WrapWriter(iw)
	return
}

func (w *writer)Close()(error){
	if c, ok := w.Writer.(io.Closer); ok {
		return c.Close()
	}
	return nil
}

func (w *writer)WriteBool(v bool)(error){
	if v {
		return w.WriteByte(1)
	}
	return w.WriteByte(0)
}

func (w *writer)WriteByte(v byte)(err error){
	if bw, ok := w.Writer.(io.ByteWriter); ok {
		return bw.WriteByte(v)
	}
	_, err = w.Write([]byte{v})
	return
}

func (w *writer)WriteUint16(v uint16)(err error){
	var buf [2]byte
	_, err = w.Write(EncodeUint16(buf[:], v))
	return
}

func (w *writer)WriteUint32(v uint32)(err error){
	var buf [4]byte
	_, err = w.Write(EncodeUint32(buf[:], v))
	return
}

func (w *writer)WriteUint64(v uint64)(err error){
	var buf [8]byte
	_, err = w.Write(EncodeUint64(buf[:], v))
	return
}

func (w *writer)WriteFloat32(v float32)(err error){
	var buf [4]byte
	_, err = w.Write(EncodeFloat32(buf[:], v))
	return
}

func (w *writer)WriteFloat64(v float64)(err error){
	var buf [8]byte
	_, err = w.Write(EncodeFloat64(buf[:], v))
	return
}

func (w *writer)WriteString(v string)(err error){
	return w.WriteBytes(([]byte)(v))
}

func (w *writer)WriteBools(v []bool)(error){
	buf := make([]byte, len(v))
	for i, ok := range v {
		if ok {
			buf[i] = 1
		}
	}
	return w.WriteBytes(buf)
}

func (w *writer)WriteBytes(v []byte)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	_, err = w.Write(v)
	return
}

func (w *writer)WriteUint16s(v []uint16)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	buf := make([]byte, len(v) * 2)
	_, err = w.Write(EncodeUint16s(buf, v))
	return
}

func (w *writer)WriteUint32s(v []uint32)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	buf := make([]byte, len(v) * 4)
	_, err = w.Write(EncodeUint32s(buf, v))
	return
}

func (w *writer)WriteUint64s(v []uint64)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	buf := make([]byte, len(v) * 8)
	_, err = w.Write(EncodeUint64s(buf, v))
	return
}

func (w *writer)WriteFloat32s(v []float32)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	buf := make([]byte, len(v) * 4)
	_, err = w.Write(EncodeFloat32s(buf, v))
	return
}

func (w *writer)WriteFloat64s(v []float64)(err error){
	if err = w.WriteUint32((uint32)(len(v))); err != nil {
		return
	}
	buf := make([]byte, len(v) * 8)
	_, err = w.Write(EncodeFloat64s(buf, v))
	return
}

