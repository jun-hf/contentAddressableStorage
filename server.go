package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
)
func init() {
	gob.Register(MessageStoreFile{})
}

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
	Key string
}

func (f *FileServer) StoreFile(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	messg := Message{
		Payload: MessageStoreFile{
			Key: key,
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
	time.Sleep(3 * time.Second)
	for _, p := range f.peers {
		if err := p.Send([]byte("This is a large file")); err != nil {
			return err
		}
	}
	return nil
}

func (f *FileServer) loop() {
	defer func() {
		f.serverTransport.Close()
		fmt.Println("Closed serverTransport")
	}()
	for {
		select {
		case remoteMessage := <-f.serverTransport.Consume():
			message := Message{}
			fmt.Println("Before decoding:")
			if err := gob.NewDecoder(bytes.NewReader(remoteMessage.Payload)).Decode(&message); err != nil {
				log.Println(err)
			}
			if err := f.handleMessage(remoteMessage.From, &message); err != nil {
				log.Println(err)
			}
		case <-f.quitCh:
			return
		}
	}
}

func (f *FileServer) handleMessage(from string, m *Message) error {
	switch payload := m.Payload.(type) {
	case MessageStoreFile:
		return f.handleMessageStoreFile(from, payload)
	}
	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, m MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		panic("peer not found")
	}
	fmt.Println(m.Key)
	if err := f.fileStorage.Write(m.Key, peer); err != nil {
		peer.(*p2p.TCPPeer).Wg.Done()
		return err
	}
	peer.(*p2p.TCPPeer).Wg.Done()
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
