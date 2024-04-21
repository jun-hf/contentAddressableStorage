package main

import (
	"fmt"

	"github.com/jun-hf/contentAddressableStorage/p2p"
	"github.com/jun-hf/contentAddressableStorage/store"
)

type FileServerOpts struct {
	fileStorageRoot   string
	transformPathFunc store.TransformPathFunc
	serverTransport   p2p.Transport
}

type FileServer struct {
	fileStorage     *store.Store
	serverTransport p2p.Transport
	quitCh chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := store.StoreOpts{
		TransformPathFunc: opts.transformPathFunc,
		Root:              opts.fileStorageRoot,
	}
	newStore := store.NewStore(storeOpts)
	return &FileServer{
		fileStorage:     newStore,
		serverTransport: opts.serverTransport,
		quitCh: make(chan struct{}),
	}
}

func(f *FileServer) Start() error {
	if err := f.serverTransport.ListenAndAccept(); err != nil {
		return err
	}
	f.Loop()
	return nil
}

func(f *FileServer) Loop() {
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

func (f *FileServer) Quit() {
	close(f.quitCh)
}