
package encoding

import (
	"bytes"
	// "io"
)

type Buffer struct{
	*bytes.Buffer
	reader
	writer
}

var _ Reader = (*Buffer)(nil)
var _ Writer = (*Buffer)(nil)

func NewBuffer(buf []byte)(b *Buffer){
	b = new(Buffer)
	b.Buffer = bytes.NewBuffer(buf)
	b.reader = reader{b.Buffer}
	b.writer = writer{b.Buffer}
	return
}

func (b *Buffer)Close()(error){
	return nil
}

func (b *Buffer)ReadByte()(v byte, err error){
	return b.reader.ReadByte()
}

func (b *Buffer)ReadString()(v string, err error){
	return b.reader.ReadString()
}

func (b *Buffer)ReadBytes()(v []byte, err error){
	return b.reader.ReadBytes()
}

func (b *Buffer)WriteByte(v byte)(err error){
	return b.writer.WriteByte(v)
}

func (b *Buffer)WriteString(v string)(err error){
	return b.writer.WriteString(v)
}
