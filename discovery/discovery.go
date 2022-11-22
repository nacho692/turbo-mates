package discovery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"sync"
)

// TODO:
// Split messages and their handlers, maybe register them in a centralized map
// Make a request-response validation in order to avoid mixing up messages
const (
	addrBytes = 8
	// alpha indicates how many neighbours to share on a lookup message
	alpha = 3
	// k is the max size of each kademlia bucket, we are unbounded atm as we
	// are not using it
	k = 5
)

// Message types
const (
	lookup  = "lookup"
	friends = "friends"
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
	Name      string
	Port      int
	Debug     bool
	ID        ID
	Bootstrap *net.UDPAddr

	mu      sync.RWMutex
	buckets [addrBytes * 8][]*peer
	socket  *net.UDPConn
}

type peer struct {
	ID   ID
	Addr *net.UDPAddr
	conn *net.UDPConn
}

func (p *peer) send(m *message) {
	b, _ := json.Marshal(&m)
	_, err := p.conn.WriteToUDP(b, p.Addr)
	if err != nil {
		fmt.Printf("[Discovery] Sending UDP: %v\n", err)
	}
}

type message struct {
	Sender  ID              `json:"sender"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type lookupPayload struct {
	ID ID
}

type friendsPayload struct {
	Friends []*peer
}

// Start starts the peer discovery, can be stopped by canceling the context.
func (d *Discovery) Start(ctx context.Context) error {

	if d.Name == "" {
		d.Name = "Discovery"
	}

	var err error
	d.socket, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   nil,
		Port: d.Port,
		Zone: "",
	})
	if err != nil {
		return err
	}
	defer d.socket.Close()

	if d.Bootstrap != nil {
		payload, _ := json.Marshal(&lookupPayload{
			ID: d.ID,
		})
		m, _ := json.Marshal(&message{
			Sender:  d.ID,
			Type:    lookup,
			Payload: payload,
		})
		_, err = d.socket.WriteToUDP(m, d.Bootstrap)
		if err != nil && d.Debug {
			fmt.Printf("[%s] Bootstrap lookup send: %v\n", d.Name, err)
		}
	}
	read := make([]byte, 2048)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, addr, err := d.socket.ReadFromUDP(read)
		if err != nil {
			if d.Debug {
				fmt.Printf("[%s] Reading UDP: %v\n", d.Name, err)
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
	p := &peer{
		ID:   m.Sender,
		Addr: addr,
		conn: d.socket,
	}
	switch m.Type {
	case lookup:
		d.handleLookup(p, m)
	case friends:
		d.handleFriends(p, m)
	}
}

func (d *Discovery) handleLookup(p *peer, m *message) {
	payload := &lookupPayload{}
	err := json.Unmarshal(m.Payload, payload)
	if err != nil {
		if d.Debug {
			fmt.Printf("[Discovery] Decoding lookup payload: %v\n", err)
		}
	}

	d.addOrRefresh(p)

	d.mu.RLock()
	var peers []*peer
	for _, b := range d.buckets {
		peers = append(peers, b...)
	}
	sort.Slice(peers, func(i, j int) bool {
		return peers[i].ID.distance(payload.ID) <= peers[i].ID.distance(payload.ID)
	})

	pLen := 0
	if len(peers) <= k {
		pLen = len(peers)
	}
	response, err := json.Marshal(&friendsPayload{
		Friends: peers[:pLen],
	})
	d.mu.RUnlock()
	if err != nil {
		if d.Debug {
			fmt.Printf("[Discovery] Encoding friends payload: %v\n", err)
		}
	}
	p.send(&message{
		Sender:  d.ID,
		Type:    friends,
		Payload: response,
	})
}

func (d *Discovery) handleFriends(p *peer, m *message) {
	payload := &friendsPayload{}
	err := json.Unmarshal(m.Payload, payload)
	if err != nil {
		if d.Debug {
			fmt.Printf("[Discovery] Decoding friends payload: %v\n", err)
		}
	}

	d.addOrRefresh(p)
	for _, f := range payload.Friends {
		d.addOrRefresh(f)
	}
}

func (d *Discovery) addOrRefresh(peer *peer) {
	idx := distanceToBucket(peer.ID.distance(d.ID))

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

func distanceToBucket(d distance) uint {
	higherEnd := distance(1) << (addrBytes*8 - 1)
	if d == 0 {
		// ourselves?
		return 0
	}
	idx := uint(addrBytes*8 - 1)
	// we count the amount of 1s until the first 0, that's our bucket
	for higherEnd&d > 0 {
		higherEnd >>= 1
		idx--
	}
	return idx
}
