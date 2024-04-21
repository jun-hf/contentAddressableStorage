package main

import (
	"log"
	"time"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
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
	fileServerOpt := FileServerOpts{
		fileStorageRoot: ":8080_directory",
		transformPathFunc: store.CASPathTransformFunc,
		serverTransport: newTCPTransport,
	}
	
	fileServer := NewFileServer(fileServerOpt)
	go func() {
		time.Sleep(5 * time.Second)
		fileServer.Quit()
	}()
	log.Fatal(fileServer.Start())
}
