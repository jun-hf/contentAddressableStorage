package p2p

import (
	"fmt"
	"log"
	"net"
)

// TCPPeer is the remote peer in the tcp transport
type TCPPeer struct {
	// connection represent the underlying connection of the peer
	connection net.Conn

	// outbound is true when TCPTransport send the connection
	// outbound is false when TCPTransport received and accepted a connection
	outbound bool
}

func (t TCPPeer) Close() error {
	return t.connection.Close()
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		connection: conn,
		outbound:   outbound,
	}
}

type TCPTransportConfig struct {
	ListenAddress string
	Decoder       Decoder
	ShakeHandFunc HandShakeFunc
	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportConfig
	listener  net.Listener
	messageCh chan Message
}

func NewTCPTransport(config TCPTransportConfig) *TCPTransport {
	return &TCPTransport{
		TCPTransportConfig: config, 
		messageCh: make(chan Message),
	}
}

func (t *TCPTransport) Consume() <-chan Message {
	return t.messageCh
}
func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddress)
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
	newTCPPeer := NewTCPPeer(conn, false)
	if err := t.ShakeHandFunc(newTCPPeer); err != nil {
		conn.Close() // reject connection when hand shake fails
		fmt.Printf("TCP error fail to validate hand shake: %v", err)
		return
	}
	fmt.Printf("Connected with %v\n", conn.RemoteAddr().String())
	msg := Message{}
	for {
		if err := t.Decoder.Decode(conn, &msg); err != nil {
			fmt.Printf("TCP error unable to decode: %v\n", err)
		}
		msg.Address = conn.RemoteAddr()
		t.messageCh <- msg
	}
}
