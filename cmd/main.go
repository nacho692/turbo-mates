package main

import (
	"context"
	"fmt"
	"github.com/nacho692/turbo-mates/discovery"
)

func main() {
	d := discovery.Discovery{
		Port:  7070,
		Debug: true,
		ID:    discovery.NewID(),
	}

	err := d.Start(context.Background())
	fmt.Println(err)
}
