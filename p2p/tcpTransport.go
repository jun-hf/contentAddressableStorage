package p2p

import (
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer is the remote peer in the tcp transport
type TCPPeer struct {
	// connection represent the underlying connection of the peer
	connection net.Conn

	// outbound is true when TCPTransport send the connection
	// outbound is false when TCPTransport received and accepted a connection
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		connection: conn,
		outbound:   outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	shakeHands    HandShakeFunc
	decoder        Decoder

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

func NewTCPTransport(address string) *TCPTransport {
	return &TCPTransport{
		listenAddress: address,
		shakeHands:    NoHandShakeFunc,
	}
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

type temp struct {}
func (t *TCPTransport) handleConnection(conn net.Conn) {
	newTCPPeer := NewTCPPeer(conn, false)
	if err := t.shakeHands(newTCPPeer); err != nil {
		conn.Close() // reject connection when hand shake fails
	}
	msg := &temp{}
	for {
		if err := t.decoder.Decode(newTCPPeer, msg); err != nil {
			fmt.Printf("TCP error unable to decode: %v\n", err)
		}
	}
}
