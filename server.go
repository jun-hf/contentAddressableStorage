package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
)

type FileServerOpts struct {
	FileStorageRoot   string
	TransformPathFunc store.TransformPathFunc
	ServerTransport   p2p.Transport
	OutboundServers   []string
}

type FileServer struct {
	fileStorage     *store.Store
	serverTransport p2p.Transport
	quitCh          chan struct{}
	outboundServers []string

	mu    sync.Mutex
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
		quitCh:          make(chan struct{}),
		outboundServers: opts.OutboundServers,
		peers:           make(map[net.Addr]p2p.Peer),
	}
}

func (f *FileServer) Start() error {
	if err := f.serverTransport.ListenAndAccept(); err != nil {
		return err
	}
	go f.dailOutbondServer()
	f.loop()
	return nil
}

func (f *FileServer) broadcast(p *Payload) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	peersList := make([]io.Writer, 0)
	for _, peer := range f.peers {
		peersList = append(peersList, peer)
	}
	mw := io.MultiWriter(peersList...)
	return gob.NewEncoder(mw).Encode(p)
}

type Payload struct {
	From string
	Data []byte
	A    A
}

type A struct {
	Inside string
}

func (f *FileServer) StoreFile(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)
	if err := f.fileStorage.Write(key, tee); err != nil {
		return err
	}
	apple := A{Inside:"Hello"}
	fmt.Println(apple)
	p := &Payload{
		From: f.serverTransport.Addr(),
		Data: buf.Bytes(),
		A:    apple,
	}
	fmt.Printf("%+v\n", p)
	return f.broadcast(p)
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.peers[p.RemoteAddr()] = p
	fmt.Printf("%v connected with %v\n", f.serverTransport.Addr(), f.peers)
	return nil
}

func (f *FileServer) Quit() {
	close(f.quitCh)
}

func (f *FileServer) dailOutbondServer() {
	for _, addr := range f.outboundServers {
		if err := f.serverTransport.Dial(addr); err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func (f *FileServer) loop() {
	defer func() {
		f.serverTransport.Close()
		fmt.Println("Closed serverTransport")
	}()
	for {
		select {
		case mes := <-f.serverTransport.Consume():
			fmt.Printf("\nMessage: %v from: %v\n", mes.Payload, mes.Address)
			pay := Payload{}
			err := gob.NewDecoder(bytes.NewReader(mes.Payload)).Decode(&pay)
			if err != nil {
				log.Print(err)
			}
			fmt.Println(string(pay.Data))
			fmt.Printf("%+v\n", pay.A.Inside)
		case <-f.quitCh:
			return
		}
	}
}
