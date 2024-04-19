package p2p

import "net"

type Message struct {
	Payload []byte
	Address net.Addr
}