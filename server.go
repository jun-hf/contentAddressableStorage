package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
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
	peers map[string]p2p.Peer
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
		peers:           make(map[string]p2p.Peer),
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

func (f *FileServer) broadcast(p *Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	peersList := make([]io.Writer, 0)
	for _, peer := range f.peers {
		peersList = append(peersList, peer)
	}
	mw := io.MultiWriter(peersList...)
	return gob.NewEncoder(mw).Encode(p)
}

type Message struct {
	Payload any
}

type MessageStoreFile struct {
	key string
}

func (f *FileServer) StoreFile(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	messg := Message{
		Payload: MessageStoreFile{
			key: key,
		},
	}
	if err := gob.NewEncoder(buf).Encode(messg); err != nil {
		return err
	}
	for _, p := range f.peers {
		if err := p.Send(buf.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileServer) OnPeer(p p2p.Peer) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.peers[p.RemoteAddr().String()] = p
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
			pay := Message{}
			fmt.Println("Before decoding:")
			err := gob.NewDecoder(bytes.NewReader(mes.Payload)).Decode(&pay)
			if err != nil {
				log.Print(err)
			}
			peer, ok := f.peers[mes.From]
			if !ok {
				panic("peer not found")
			}
			fmt.Printf("The message: %s\n", string(pay.Payload.([]byte)))
			buf := make([]byte, 1000)
			n, err := peer.Read(buf)
			fmt.Print(string(buf[:n]))
			if err != nil {
				panic(err)
			}
			peer.(*p2p.TCPPeer).Wg.Done()
		case <-f.quitCh:
			return
		}
	}
}
