package main

import (
	"github.com/nacho692/turbo-mates/discovery"
	"math/rand"
)

// main to test bucketing
func main() {
	rand.Seed(1)
	for i := 0; i < 10; i++ {
		d := discovery.Discovery{
			Port:  7070,
			Debug: true,
			ID:    discovery.NewID(),
		}
		d.Send()
	}
}
