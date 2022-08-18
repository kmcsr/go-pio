
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
)

var _ PacketAsk = (*Ping)(nil)
var _ PacketBase = (*Pong)(nil)

func (*Ping)PktId()(uint32){ return 0x01 }
func (*Pong)PktId()(uint32){ return 0x02 }

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

