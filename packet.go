
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

