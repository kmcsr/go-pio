
package pio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"github.com/kmcsr/go-pio/encoding"
)


const (
	NoAsk   byte = 0x00
	SendAsk byte = 0x01
	RecvAsk byte = 0x02
)

type Conn struct{
	r encoding.Reader
	w encoding.Writer

	started bool
	err error
	ctx context.Context
	cancel context.CancelFunc

	wmux sync.Mutex
	idmux sync.Mutex
	idins uint32
	waits map[uint32]chan PacketBase

	pkts map[uint32]PacketNewer
}

func NewConn(r io.Reader, w io.Writer)(c *Conn){
	return NewConnContext(context.Background(), r, w)
}

func NewConnContext(ctx context.Context, r io.Reader, w io.Writer)(c *Conn){
	ctx, cancel := context.WithCancel(ctx)
	c = &Conn{
		r: encoding.WrapReader(r),
		w: encoding.WrapWriter(w),
		started: false,
		ctx: ctx,
		cancel: cancel,
		waits: make(map[uint32]chan PacketBase),
		pkts: make(map[uint32]PacketNewer),
	}
	c.initPkts()
	return
}

func Pipe()(a, b *Conn){
	ar, bw := io.Pipe()
	br, aw := io.Pipe()
	a, b = NewConn(ar, aw), NewConn(br, bw)
	return
}

func OsPipe()(c *Conn, br, bw *os.File, err error){
	ar, bw, err := os.Pipe()
	if err != nil {
		return
	}
	br, aw, err := os.Pipe()
	if err != nil {
		return
	}
	c = NewConn(ar, aw)
	return
}

func (c *Conn)Close()(err error){
	err = c.r.Close()
	err2 := c.w.Close()
	if err == nil {
		err = err2
	}
	return
}

func (c *Conn)Context()(context.Context){
	return c.ctx
}

func (c *Conn)Err()(error){
	return c.err
}

func (c *Conn)initPkts(){
	c.AddPacket(func()(PacketBase){ return new(Ping) })
	c.AddPacket(func()(PacketBase){ return new(Pong) })
}

func (c *Conn)AddPacket(newer PacketNewer){
	if newer == nil {
		panic("newer cannot be nil")
	}
	sample := newer()
	pid := sample.PktId()
	if _, ok := c.pkts[pid]; ok {
		panic("Packet id already exists")
	}
	c.pkts[pid] = newer
}

func (c *Conn)NewPacket(pid uint32)(PacketBase){
	if newer, ok := c.pkts[pid]; ok {
		return newer()
	}
	return nil
}

func (c *Conn)Ping()(ping time.Duration, err error){
	return c.PingWith(context.Background())
}

func (c *Conn)PingWith(ctx context.Context)(ping time.Duration, err error){
	begin := time.Now()
	pay := (uint64)(begin.Unix() * 1000 + begin.UnixNano() / 1000000 % 1000)
	var res PacketBase
	if res, err = c.AskWith(ctx, &Ping{pay}); err != nil {
		return
	}
	if pay != res.(*Pong).Payload {
		err = errors.New("ping-pong payload not same")
		return
	}
	ping = time.Since(begin)
	return
}

func (c *Conn)Send(p PacketBase)(err error){
	return c.send(p, 0, NoAsk)
}

func (c *Conn)Ask(p PacketBase)(res PacketBase, err error){
	return c.AskWith(context.Background(), p)
}

func (c *Conn)AskWith(ctx context.Context, p PacketBase)(res PacketBase, err error){
	ch := make(chan PacketBase, 1)
	c.idmux.Lock()
	c.idins++
	id := c.idins
	for {
		_, ok := c.waits[id]
		if !ok {
			break
		}
		id++
	}
	c.waits[id] = ch
	defer delete(c.waits, id)
	c.idmux.Unlock()
	if err = c.send(p, id, SendAsk); err != nil {
		return
	}
	select {
	case res = <- ch:
		return
	case <-ctx.Done():
		err = ctx.Err()
		return
	case <-c.ctx.Done():
		err = c.ctx.Err()
		return
	}
}

func (c *Conn)send(p PacketBase, id uint32, ask byte)(err error){
	if ask == NoAsk {
		id = 0
	}
	buf := bytes.NewBuffer(nil)
	wr := encoding.WrapWriter(buf)
	wr.WriteUint32(id)
	wr.WriteByte(ask)
	wr.WriteUint32(p.PktId())
	if err = p.WriteTo(wr); err != nil {
		return
	}

	c.wmux.Lock()
	defer c.wmux.Unlock()

	if err = c.w.WriteBytes(buf.Bytes()); err != nil {
		return
	}
	return
}

func (c *Conn)parser(buf []byte){
	var (
		id uint32
		ask byte
		pid uint32
		p PacketBase
		err error
	)

	rd := encoding.WrapReader(bytes.NewReader(buf))
	if id, err = rd.ReadUint32(); err != nil {
		return
	}
	if ask, err = rd.ReadByte(); err != nil {
		return
	}
	if pid, err = rd.ReadUint32(); err != nil {
		return
	}
	p = c.NewPacket(pid)
	if p == nil {
		return
	}
	err = p.ParseFrom(rd)
	if err != nil {
		return
	}
	switch ask {
	case SendAsk:
		if pa, ok := p.(PacketAsk); ok {
			var rv PacketBase
			if rv, err = pa.Ask(); err != nil {
				return
			}
			if rv == nil {
				rv = OkPkt
			}
			if err = c.send(rv, id, RecvAsk); err != nil {
				return
			}
		}
	case RecvAsk:
		c.idmux.Lock()
		ch, ok := c.waits[id]
		if ok {
			delete(c.waits, id)
		}
		c.idmux.Unlock()
		if ok {
			ch <- p
		}
		return
	case NoAsk:
		if pa, ok := p.(Packet); ok {
			if err = pa.Trigger(); err != nil {
				return
			}
		}
	default:
		panic("Unknown ask value")
		return
	}
}

func (c *Conn)Serve()(err error){
	if c.started {
		panic("Conn already served")
	}
	c.started = true
	var buf []byte
	defer c.cancel()
	for {
		if buf, err = c.r.ReadBytes(); err != nil {
			break
		}
		c.parser(buf)
	}
	if err != nil {
		c.err = err
	}
	return
}
