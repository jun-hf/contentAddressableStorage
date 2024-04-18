package main

import (
	"log"
	"github.com/jun-hf/contentAddressableStorage/p2p"
)

func main() {
	config := p2p.TCPTransportConfig{
		ListenAddress: "localhost:8080",
		Decoder: p2p.GOBDecoder{},
		ShakeHandFunc: p2p.NoHandShakeFunc,
	}
	newTCPTransport := p2p.NewTCPTransport(config)
	if err := newTCPTransport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}