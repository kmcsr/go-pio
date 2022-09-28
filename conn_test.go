
package pio_test

import (
	"testing"

	// "github.com/kmcsr/go-pio/encoding"
	. "github.com/kmcsr/go-pio"
)

func TestConnPing(t *testing.T){
	c, d := Pipe()
	go d.Serve()
	go c.Serve()
	defer c.Close()

	<-c.ServeDone()
	ping, err := c.Ping()
	if err != nil {
		t.Fatalf("Ping: %v", err)
	}
	t.Logf("ping: %v", ping)
}

func TestConnAsStream(t *testing.T){
	c, d := Pipe()
	go d.Serve()
	go c.Serve()
	defer c.Close()
	defer d.Close()

	<-c.ServeDone()
	if _, err := c.Ping(); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	crw, err := c.AsStream()
	if err != nil {
		t.Fatalf("c.AsStream: %v", err)
	}
	go func(){
		t.Logf("Writing")
		crw.Write(([]byte)("hello pio"))
	}()

	<-d.StreamedDone()
	drw, err := d.AsStream()
	if err != nil {
		t.Fatalf("d.AsStream: %v", err)
	}
	t.Logf("Reading")
	var buf = make([]byte, 8)
	n, err := drw.Read(buf)
	if err != nil {
		t.Fatalf("d.stream.Read: %v", err)
	}
	t.Logf("Readed: '%v'", (string)(buf[:n]))
}
