package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	connection net.Conn

	// inbound is true when TCPTransport accept the connection
	// inbound is false when TCPTransport send a connection
	inbound bool
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{listenAddress: address}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}
	log.Printf("Starting to listen at %v\n", t.listener.Addr().String())
	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go t.handleConnection(conn)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn) {
	newTCPPeer := TCPPeer{ connection: conn, inbound: true}
	fmt.Fprint(conn, "Hello!!")
}