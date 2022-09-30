
package pio

import (
	"github.com/kmcsr/go-pio/encoding"
)

type (
	Ping struct{
		Payload uint64
	}
	Pong struct{
		Payload uint64
	}
	Ok struct{}
)

var _ PacketAsk = (*Ping)(nil)
var _ PacketBase = (*Pong)(nil)
var OkPkt PacketBase = Ok{}

func (*Ping)PktId()(uint32){ return 0x01 }
func (*Pong)PktId()(uint32){ return 0x02 }
func (Ok)PktId()(uint32){ return 0x04 }

func (p *Ping)ParseFrom(r encoding.Reader)(err error){
	p.Payload, err = r.ReadUint64()
	return
}

func (p *Ping)WriteTo(w encoding.Writer)(err error){
	err = w.WriteUint64(p.Payload)
	return
}

func (p *Ping)Ask()(res PacketBase, err error){
	res = &Pong{
		Payload: p.Payload,
	}
	return
}

func (p *Pong)ParseFrom(r encoding.Reader)(err error){
	p.Payload, err = r.ReadUint64()
	return
}

func (p *Pong)WriteTo(w encoding.Writer)(err error){
	err = w.WriteUint64(p.Payload)
	return
}

func (Ok)WriteTo(encoding.Writer)(error){ return nil }
func (Ok)ParseFrom(encoding.Reader)(error){ return nil }

type streamingError struct{}

var streamingErr error = streamingError{}

func (streamingError)Error()(string){
	return "As streaming"
}

var (
	stmPing PacketBase = NewPkt(0x10)
	stmPong PacketBase = NewPkt(0x11)
)
