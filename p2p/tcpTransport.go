package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
)

// TCPPeer is the remote peer in the tcp transport
type TCPPeer struct {
	// net.Conn represent the underlying connection 
	// which in this case is a tcp
	net.Conn

	// outbound is true when TCPTransport send the connection
	// outbound is false when TCPTransport received and accepted a connection
	outbound bool
}

func (t *TCPPeer) Send(b []byte) error {
	_, err := t.Write(b)
	return err
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn: conn,
		outbound:   outbound,
	}
}

type TCPTransportConfig struct {
	ListenAddress string
	Decoder       Decoder
	ShakeHandFunc HandShakeFunc
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportConfig
	listener  net.Listener
	messageCh chan Message
}

func NewTCPTransport(config TCPTransportConfig) *TCPTransport {
	return &TCPTransport{
		TCPTransportConfig: config,
		messageCh:          make(chan Message),
	}
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Consume() <-chan Message {
	return t.messageCh
}

func (t *TCPTransport) Addr() string {
	return t.listener.Addr().String()
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

func (t *TCPTransport) Dial(listenAddr string) error {
	conn, err := net.Dial("tcp", listenAddr)
	if err != nil {
		return err
	}
	go t.handleConnection(conn, true)
	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			log.Print(err)
			continue
		}
		go t.handleConnection(conn, false)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, outbound bool) {
	var err error
	defer func() {
		conn.Close()
		fmt.Printf("TCP error dropping connection: %v\n", err)
	}()
	newTCPPeer := NewTCPPeer(conn, outbound)
	if err = t.ShakeHandFunc(newTCPPeer); err != nil {
		err = fmt.Errorf("failed hand shake: %v", err)
		return
	}
	if t.OnPeer != nil {
		if err = t.OnPeer(newTCPPeer); err != nil {
			err = fmt.Errorf("failed OnPeer: %v", err)
			return
		}
	}
	fmt.Printf("Connected with %v\n", conn.RemoteAddr().String())
	msg := Message{}
	for {
		if err := t.Decoder.Decode(conn, &msg); err != nil {
			fmt.Printf("TCP error: at handleConnection unable to decode: %v\n", err)
			return
		}
		msg.Address = conn.RemoteAddr()
		t.messageCh <- msg
	}
}
