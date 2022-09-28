
package pio

import (
	"github.com/kmcsr/go-pio/encoding"
)

type (
	PacketBase interface{
		PktId()(uint32)
		ParseFrom(encoding.Reader)(error)
		WriteTo(encoding.Writer)(error)
	}

	Packet interface{
		PacketBase
		Trigger()(error)
	}

	PacketAsk interface{
		PacketBase
		Ask()(PacketBase, error)
	}
)

type PacketNewer func()(PacketBase)

type (
	EmptyPkt struct{
		Id uint32
	}
	EmptyPktTrigger struct{
		EmptyPkt
		OnTrigger func()(error)
	}
	EmptyPktAsk struct{
		EmptyPkt
		OnAsk func()(PacketBase, error)
	}
)

var _ PacketBase = EmptyPkt{}
var _ Packet     = EmptyPktTrigger{}
var _ PacketAsk  = EmptyPktAsk{}

func NewPkt(id uint32)(PacketBase){
	return EmptyPkt{
		Id: id,
	}
}

func NewPktTrigger(id uint32, ontrigger func()(error))(Packet){
	return EmptyPktTrigger{
		EmptyPkt: EmptyPkt{id},
		OnTrigger: ontrigger,
	}
}

func NewPktAsk(id uint32, onask func()(PacketBase, error))(PacketAsk){
	return EmptyPktAsk{
		EmptyPkt: EmptyPkt{id},
		OnAsk: onask,
	}
}

func (pkt EmptyPkt)PktId()(uint32){ return pkt.Id }
func (EmptyPkt)ParseFrom(encoding.Reader)(error){ return nil }
func (EmptyPkt)WriteTo(encoding.Writer)(error){ return nil }
func (pkt EmptyPktTrigger)Trigger()(error){
	return pkt.OnTrigger()
}
func (pkt EmptyPktAsk)Ask()(PacketBase, error){
	if pkt.OnAsk == nil {
		panic("pkt.OnAsk == nil")
	}
	return pkt.OnAsk()
}

