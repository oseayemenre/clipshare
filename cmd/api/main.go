package main

import (
	"github.com/oseayemenre/clip_share/internal/ip"
	"log"
)

func main() {
	svr := NewServer(ip.GetPrivateIp())

	if err := svr.run(); err != nil {
		log.Fatal(err)
	}
}
