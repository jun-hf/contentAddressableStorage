package main

import (
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
)

type FileServerOpts struct {
	FileStorageRoot   string
	TransformPathFunc store.TransformPathFunc
	ServerTransport   p2p.Transport
	OutboundServers []string
}

type FileServer struct {
	fileStorage     *store.Store
	serverTransport p2p.Transport
	quitCh chan struct{}
	outboundServers []string

	mu sync.Mutex
	peers map[net.Addr]p2p.Peer
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		TransformPathFunc: opts.TransformPathFunc,
		Root:              opts.FileStorageRoot,
	}
	newStore := store.NewStore(storeOpts)
	return &FileServer{
		fileStorage:     newStore,
		serverTransport: opts.ServerTransport,
		quitCh: make(chan struct{}),
		outboundServers: opts.OutboundServers,
		peers: make(map[net.Addr]p2p.Peer),
	}
}

func(f *FileServer) Start() error {
	if err := f.serverTransport.ListenAndAccept(); err != nil {
		return err
	}
	go f.dailOutbondServer()
	f.loop()
	return nil
}
type payload struct {}

func (f *FileServer) broadcast(p payload) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	peerList := []io.Writer{}
	for _, peer := range f.peers {
		peerList = append(peerList, peer)
	}
	mw := io.MultiWriter(peerList...)
	return gob.NewEncoder(mw).Encode(p)
}

func (f *FileServer) StoreFile(key string, r io.Reader) error {
	return nil
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.peers[p.RemoteAddr()] = p
	fmt.Println(f.peers)
	return nil
}

func (f *FileServer) Quit() {
	close(f.quitCh)
}

func (f *FileServer) dailOutbondServer() {
	fmt.Println(f.outboundServers)
	for _, addr := range f.outboundServers {
		if err := f.serverTransport.Dial(addr); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func(f *FileServer) loop() {
	defer func() {
		f.serverTransport.Close()
		fmt.Println("Closed serverTransport")
	} ()
	for {
		select {
		case mes := <- f.serverTransport.Consume():
			fmt.Println(mes)
		case <- f.quitCh:
			return
		}
	}
}