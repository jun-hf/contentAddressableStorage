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

func (f *FileServer) broadcast(m *Message) error {
	buff := new(bytes.Buffer)
	if err := gob.NewEncoder(buff).Encode(m); err != nil {
		return err
	}
	for _, peer := range f.peers {
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buff.Bytes()); err != nil {
			return err
		}
	}
	return nil
}

type Message struct {
	Payload any
}

type MessageGetFile struct {
	Key string
}

type MessageStoreFile struct {
	Key string
	Size int64
}

func (f *FileServer) Get(key string) (io.Reader, error) {
	if f.fileStorage.Has(key) {
		return f.fileStorage.Read(key)
	}

	message := &Message{
		Payload: MessageGetFile{
			Key: key,
		},
	}
	log.Printf("key: %v does not found locally broadcasting to remote peer", key)
	if err := f.broadcast(message); err != nil {
		return nil, err
	}
	return nil, nil
}
func (f *FileServer) Store(key string, r io.Reader) error {
	messageData := new(bytes.Buffer)
	tee := io.TeeReader(r, messageData)
	n, err := f.fileStorage.Write(key, tee)
	if err != nil {
		return err
	}
	message := &Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: n,
		},
	}
	if err := f.broadcast(message); err != nil {
		return err
	}
	time.Sleep(3 *time.Second)
	// please write this to a multi writer
	for _, peer := range f.peers {
		peer.Send([]byte{p2p.IncomingStream})
		n, err := io.Copy(peer, messageData)
		if err != nil {
			log.Printf("Write to peer (%v) failed: %v\n", peer.RemoteAddr().String(), err)
		}
		fmt.Printf("Written %v bytes to peer (%v)\n", n, peer.RemoteAddr().String())
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
				log.Println("Decode failed ", err)
			}
			if err := f.handleMessage(remoteMessage.From, &message); err != nil {
				log.Println("handleMessage failed ", err)
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
	case MessageGetFile:
		return f.handleMessageGetFile(from, payload)
	}
	return nil
}

func (f *FileServer) handleMessageGetFile(from string, m MessageGetFile) error {
	if !f.fileStorage.Has(m.Key) {
		return fmt.Errorf("suppose to server: %v but file not exist in server", m.Key)
	}
	r, err := f.fileStorage.Read(m.Key)
	if err != nil {
		return err
	}
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("peer (%v) is not found",from)
	}
	_, err = io.Copy(peer, r)
	if err != nil {
		return fmt.Errorf("copy faile: %v", err)
	}

	return nil
}

func (f *FileServer) handleMessageStoreFile(from string, m MessageStoreFile) error {
	peer, ok := f.peers[from]
	if !ok {
		return fmt.Errorf("cannot peer (%v) not found in peer list", from)
	}
	fmt.Println("Storing file")
	fmt.Println(m.Size)
	if _, err := f.fileStorage.Write(m.Key, io.LimitReader(peer, m.Size)); err != nil {
		return err
	}
	fmt.Println("Done storing file")
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

func (f *FileServer) stream(p *Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	peersList := make([]io.Writer, 0)
	for _, peer := range f.peers {
		peersList = append(peersList, peer)
	}
	mw := io.MultiWriter(peersList...)
	return gob.NewEncoder(mw).Encode(p)
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}