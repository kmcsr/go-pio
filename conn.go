
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

type ConnState int
const (
	ConnInited ConnState = iota
	ConnServing
	ConnPreStream
	ConnStreamed
)

type Conn struct{
	r encoding.Reader
	w encoding.Writer

	status ConnState
	statusmux sync.RWMutex
	served chan struct{}
	streamed chan struct{}
	ctx context.Context
	cancel context.CancelFunc

	wmux sync.Mutex
	idmux sync.Mutex
	idins uint32
	waits map[uint32]chan PacketBase

	pkts map[uint32]PacketNewer

	OnPktNotFound func(id uint32, body encoding.Reader)
	OnParseError func(pkt PacketBase, err error)
}

func NewConn(r io.Reader, w io.Writer)(c *Conn){
	return NewConnContext(context.Background(), r, w)
}

func NewConnContext(ctx context.Context, r io.Reader, w io.Writer)(c *Conn){
	ctx, cancel := context.WithCancel(ctx)
	c = &Conn{
		r: encoding.WrapReader(r),
		w: encoding.WrapWriter(w),
		status: ConnInited,
		served: make(chan struct{}, 0),
		streamed: make(chan struct{}, 0),
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

func (c *Conn)ServeDone()(<-chan struct{}){
	return c.served
}

func (c *Conn)initPkts(){
	c.AddPacket(func()(PacketBase){ return new(Ping) })
	c.AddPacket(func()(PacketBase){ return new(Pong) })
	c.AddPacket(func()(PacketBase){ return OkPkt })
	c.AddPacket(func()(PacketBase){ return stmPing })
	c.AddPacket(func()(PacketBase){ return stmPong })
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

func (c *Conn)checkStreamed(){
	c.statusmux.RLock()
	if c.status > ConnServing {
		c.statusmux.RUnlock()
		panic("pio.Conn is streamed")
	}
	c.statusmux.RUnlock()
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
	c.checkStreamed()
	return c.send(p, 0, NoAsk)
}

func (c *Conn)Ask(p PacketBase)(res PacketBase, err error){
	return c.AskWith(context.Background(), p)
}

func (c *Conn)AskWith(ctx context.Context, p PacketBase)(res PacketBase, err error){
	c.checkStreamed()

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

func (c *Conn)parser(buf []byte)(err error){
	var (
		id uint32
		ask byte
		pid uint32
		p PacketBase
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
		if c.OnPktNotFound != nil {
			c.OnPktNotFound(pid, rd)
		}
		return
	}
	err = p.ParseFrom(rd)
	if err != nil {
		if c.OnParseError != nil {
			c.OnParseError(p, err)
		}
		return
	}
	switch ask {
	case SendAsk:
		pa := p.(PacketAsk)
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
	case NoAsk:
		if p == stmPing {
			if err = c.send(stmPong, 0, NoAsk); err != nil {
				return
			}
			return streamingErr
		}
		if p == stmPong {
			return streamingErr
		}
		if pa, ok := p.(Packet); ok {
			if err = pa.Trigger(); err != nil {
				return
			}
		}
	default:
		panic("Unexpected ask mask")
	}
	return
}

func (c *Conn)Serve()(err error){
	c.statusmux.Lock()
	if c.status != ConnInited {
		c.statusmux.Unlock()
		panic("pio.Conn already served")
	}
	c.status = ConnServing
	c.statusmux.Unlock()
	close(c.served)

	var buf []byte
	defer c.cancel()
	for {
		{
			c.statusmux.RLock()
			st := c.status
			c.statusmux.RUnlock()
			if st != ConnServing && st != ConnPreStream {
				break
			}
		}
		if buf, err = c.r.ReadBytes(); err != nil {
			return
		}
		if er := c.parser(buf); er != nil {
			if er == streamingErr {
				c.statusmux.Lock()
				c.status = ConnStreamed
				c.statusmux.Unlock()
				close(c.streamed)
				return
			}
		}
	}
	return
}

type readWriteCloser struct{
	c *Conn
}

var _ io.ReadWriteCloser = readWriteCloser{}

func (c readWriteCloser)Read(buf []byte)(n int, err error){
	return c.c.r.Read(buf)
}

func (c readWriteCloser)Write(buf []byte)(n int, err error){
	return c.c.w.Write(buf)
}

func (c readWriteCloser)Close()(err error){
	return c.c.Close()
}

func (c *Conn)StreamedDone()(<-chan struct{}){
	return c.streamed
}

func (c *Conn)AsStream()(rw io.ReadWriteCloser, err error){
	c.statusmux.Lock()
	streaming := c.status != ConnStreamed
	if streaming {
		if c.status != ConnServing {
			c.statusmux.Unlock()
			panic("pio.Conn is not serving")
		}
		c.status = ConnPreStream
		if err = c.send(stmPing, 0, NoAsk); err != nil {
			c.status = ConnServing
			c.statusmux.Unlock()
			return
		}
	}
	c.statusmux.Unlock()

	if streaming {
		select{
		case <-c.streamed:
		case <-c.ctx.Done():
			err = c.ctx.Err()
			return
		}
	}
	rw = readWriteCloser{c}
	return
}
