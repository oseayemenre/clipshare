package ip

import (
	"log"
	"net"
)

func GetPrivateIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	localAddr := conn.LocalAddr().String()

	return localAddr
}
