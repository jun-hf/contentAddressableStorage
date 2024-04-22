package p2p

import "net"

// Peer represent the remote node in the network
type Peer interface {
	net.Conn
	Send([]byte) error
}

// Transport handles any communications between peers
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan Message
	Close() error
	Dial(string) error
	Addr() string
}
