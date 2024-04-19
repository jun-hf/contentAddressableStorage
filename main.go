package main

import (
	"fmt"
	"log"

	"github.com/jun-hf/contentAddressableStorage/p2p"
)
func OnPeer(p p2p.Peer) error {
	log.Println("Getting Peer in OnPeer")
	p.Close()
	return nil
}

func main() {
	config := p2p.TCPTransportConfig{
		ListenAddress: "localhost:8080",
		Decoder:       p2p.DefaultDecoder{},
		ShakeHandFunc: p2p.NoHandShakeFunc,
		OnPeer: OnPeer,
	}
	newTCPTransport := p2p.NewTCPTransport(config)
	go func() {
		for {
			msg := <-newTCPTransport.Consume()
			fmt.Printf("Client message: %+v\n", msg)
		}
	}()
	if err := newTCPTransport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}
	select {}
}
