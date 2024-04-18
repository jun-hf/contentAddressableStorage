package main

import (
	"log"
	"github.com/jun-hf/contentAddressableStorage/p2p"
)

func main() {
	newTCPTransport := p2p.NewTCPTransport(":8080")
	if err := newTCPTransport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}