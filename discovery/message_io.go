package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"syscall"
)

func (d *Discovery) enqueue(data *bytes.Buffer, addr *net.UDPAddr) {
	select {
	case d.receivedMessages <- &receivedMessage{
		data: data,
		addr: addr,
	}:
	default:
		fmt.Printf("[%s] Discarding received message due to full queue", d.Name)
	}
}

func (d *Discovery) readMessage(data *bytes.Buffer, addr *net.UDPAddr) {
	m := &message{}
	err := json.NewDecoder(data).Decode(m)
	if err != nil {
		if d.Debug {
			fmt.Printf("[%s] Decoding metadata: %v\n", d.Name, err)
		}
	}
	p := &peer{
		ID:   m.Sender,
		Addr: addr,
	}
	switch m.Type {
	case lookup:
		d.handleLookup(p, m)
	case friends:
		d.handleFriends(p, m)
	}
}

func (d *Discovery) startReceiver(ctx context.Context) func() {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan bool)
	read := make([]byte, 2048)

	go func() {
		<-ctx.Done()
	}()
	go func() {
		for {
			n, addr, err := d.socket.ReadFromUDP(read)
			if err != nil {
				switch {
				case errors.Is(err, net.ErrClosed),
					errors.Is(err, io.EOF),
					errors.Is(err, syscall.EPIPE):
					done <- true
					return
				}
				if d.Debug {
					fmt.Printf("[%s] Reading UDP: %v\n", d.Name, err)
				}
			}
			dup := append([]byte(nil), read[:n]...)
			d.enqueue(bytes.NewBuffer(dup), addr)
		}
	}()
	return func() {
		cancel()
		<-done
		fmt.Printf("[%s] Shutting down receiver\n", d.Name)
	}
}

func (d *Discovery) startConsumer(ctx context.Context) func() {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan bool)
	go func() {
		for {
			select {
			case m := <-d.receivedMessages:
				d.readMessage(m.data, m.addr)
			case <-ctx.Done():
				done <- true
			}
		}
	}()
	return func() {
		cancel()
		<-done
		fmt.Printf("[%s] Shutting down consumer\n", d.Name)
	}
}
