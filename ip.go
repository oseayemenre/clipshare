package main

import (
	"log"
	"net"
)

func getPrivateIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().String()

	return localAddr
}
