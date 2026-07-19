package main

import "log"

func main() {
	server, err := NewServer()
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
