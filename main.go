package main

import (
	"log"
)

func main() {
	svr := NewServer(getPrivateIp())

	if err := svr.run(); err != nil {
		log.Fatal(err)
	}
}
