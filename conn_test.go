
package pio_test

import (
	"testing"

	// "github.com/kmcsr/go-pio/encoding"
	. "github.com/kmcsr/go-pio"
)

func TestConnPing(t *testing.T){
	c, d := Pipe()
	go d.Serve()
	go func(){
		defer c.Close()
		ping, err := c.Ping()
		if err != nil {
			panic(err)
		}
		t.Logf("ping: %v", ping)
	}()
	c.Serve()
}
