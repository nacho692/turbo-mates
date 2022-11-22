package discovery

import (
	"encoding/hex"
	"net"
)

type Stats struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	PeersCount int                         `json:"peers_count"`
	Port       int                         `json:"port"`
	Buckets    [addrBytes * 8][]*PeerStats `json:"buckets"`
}

type PeerStats struct {
	ID   string   `json:"id"`
	Addr net.Addr `json:"addr"`
}

func (d *Discovery) Stats() *Stats {

	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := &Stats{
		ID:   hex.EncodeToString(d.ID[:]),
		Name: d.Name,
		Port: d.Port,
	}
	for i, _ := range d.buckets {
		for _, p := range d.buckets[i] {
			stats.Buckets[i] = append(stats.Buckets[i], &PeerStats{
				ID:   hex.EncodeToString(p.ID[:]),
				Addr: p.Addr,
			})
			stats.PeersCount++
		}
	}
	return stats
}
