package p2p

// Peer represent the remote node in the network
type Peer interface {
	Close() error
}

// Transport handles any communications between peers
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan Message
	Close() error
}
