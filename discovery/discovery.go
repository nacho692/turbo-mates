package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sync"
)

const (
	addrBytes = 8
	// k is the max size of each kademlia bucket, we are unbounded atm as we
	// are not using it
	k = 5
)

// distance addressability must be addrBytes * 8
type distance uint64

type ID [addrBytes]byte

func NewID() ID {
	id := ID{}
	for i := range id {
		id[i] = uint8(rand.Uint32())
	}
	return id
}

func (u ID) distance(v ID) distance {
	dist := distance(0)
	for i := range u {
		// shift 1 byte
		dist <<= 8
		// set the lowest part to the xor result
		dist |= distance(u[i] ^ v[i])
	}
	return dist
}

type Discovery struct {
	Port  int
	Debug bool
	ID    ID

	mu      sync.Mutex
	buckets [addrBytes * 8][]*Peer
}

type Peer struct {
	ID   ID
	Addr net.Addr
}

type message struct {
	ID      ID              `json:"ID"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Start starts the peer discovery, can be stopped by canceling the context.
func (d *Discovery) Start(ctx context.Context) error {
	ln, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: d.Port,
		Zone: "",
	})
	if err != nil {
		return err
	}
	read := make([]byte, 2048)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, addr, err := ln.ReadFromUDP(read)
		if err != nil {
			if d.Debug {
				fmt.Printf("[Discovery] Reading UDP: %v\n", err)
			}
			continue
		}
		dup := append([]byte(nil), read[:n]...)
		go d.readMessage(bytes.NewBuffer(dup), addr)
	}

	return nil
}

func (d *Discovery) readMessage(data *bytes.Buffer, addr *net.UDPAddr) {
	m := &message{}
	err := json.NewDecoder(data).Decode(m)
	if err != nil {
		if d.Debug {
			fmt.Printf("[Discovery] Decoding metadata: %v\n", err)
		}
	}
	switch m.Type {
	case "Hi!":
		d.handleHi(m, addr)
	case "Ho!":
		//d.handleHo(data, addr)
	}
}

func (d *Discovery) handleHi(m *message, addr *net.UDPAddr) {
	d.addOrRefresh(&Peer{
		ID:   m.ID,
		Addr: addr,
	})
	// should return Ho!
}

func (d *Discovery) addOrRefresh(peer *Peer) {
	higherEnd := distance(1) << (addrBytes*8 - 1)
	dist := peer.ID.distance(d.ID)
	if dist == 0 {
		// ourselves?
		return
	}
	idx := addrBytes*8 - 1
	// we count the amount of 1s until the first 0, that's our bucket
	for higherEnd&dist > 0 {
		higherEnd >>= 1
		idx--
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, p := range d.buckets[idx] {
		if p.ID == peer.ID {
			// should refresh LRU
			return
		}
	}
	d.buckets[idx] = append(d.buckets[idx], peer)
}

func (d *Discovery) Send() {
	udp, _ := net.DialUDP("udp", nil, &net.UDPAddr{
		IP:   net.ParseIP("localhost"),
		Port: d.Port,
		Zone: "",
	})
	m := &message{
		ID:      d.ID,
		Type:    "Hi!",
		Payload: nil,
	}

	json.NewEncoder(udp).Encode(m)
}
