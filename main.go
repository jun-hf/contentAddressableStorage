package main

import (
	"fmt"
	"log"
	"strings"
	"time"

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
	time.Sleep(3 * time.Second)
	
	go func() {
		log.Fatal(s2.Start())
	} ()

	time.Sleep(3 * time.Second)
	if err := s2.Store("I am here", strings.NewReader("inside file")); err != nil {
		log.Printf("s2 StoreFile failed: %+v\n",err)
	}
	// _, err := s2.Get("I am here")
	// if err != nil {
	// 	log.Println(err)
	// }
	// res, err := io.ReadAll(r)
	// if err != nil {
	// 	log.Println(err)
	// }
	// fmt.Println(string(res))
	select {}
}
