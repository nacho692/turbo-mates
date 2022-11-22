package main

import (
	"context"
	"fmt"
	"github.com/nacho692/turbo-mates/discovery"
	"net/http"
)

func main() {
	d := &discovery.Discovery{
		Port:  7070,
		Debug: true,
		ID:    discovery.NewID(),
	}

	go func() {
		err := d.Start(context.Background())
		fmt.Println(err)
	}()

	http.Handle("/", discovery.Handler(d))
	_ = http.ListenAndServe(":6969", nil)
}
