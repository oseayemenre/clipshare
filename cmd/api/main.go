package main

import "log"

func main() {
	svr := NewServer("3000")

	if err := svr.run(); err != nil {
		log.Fatal(err)
	}
}
