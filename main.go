package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/nacho692/turbo-mates/discovery"
)

// main to test bucketing
func main() {
	var nodes []*discovery.Discovery
	for i := 0; i < 3; i++ {
		nodes = append(nodes, &discovery.Discovery{
			Name:  fmt.Sprintf("%d", i),
			Port:  7070 + i + 1,
			Debug: true,
			ID:    discovery.NewID(),
			Bootstrap: &net.UDPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 7070,
			},
		})
	}

	for _, n := range nodes {
		n := n
		go func() {
			err := n.Start(context.Background())
			fmt.Println(err)
		}()
	}
	time.Sleep(10 * time.Hour)
}
