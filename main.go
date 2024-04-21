package main

import (
	"fmt"
	"log"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
)

func createServer(listenAddr string, ports... string) *FileServer {
	config := p2p.TCPTransportConfig{
		ListenAddress: listenAddr,
		Decoder:       p2p.DefaultDecoder{},
		ShakeHandFunc: p2p.NoHandShakeFunc,
	}
	newTCPTransport := p2p.NewTCPTransport(config)
	fileServerOpt := FileServerOpts{
		FileStorageRoot: fmt.Sprintf("%sNetworkDir", listenAddr),
		TransformPathFunc: store.CASPathTransformFunc,
		ServerTransport: newTCPTransport,
		OutboundServers: ports,
	}

	s := NewFileServer(fileServerOpt)
	newTCPTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s := createServer(":8080")
	s2 := createServer(":3000", ":8080")

	go func() {
		log.Fatal(s.Start())
	}()
	log.Fatal(s2.Start())
}
