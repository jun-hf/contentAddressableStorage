package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestTCPTransport(t *testing.T) {
	config := TCPTransportConfig{
		ListenAddress: "localhost:8080",
		Decoder:       DefaultDecoder{},
		ShakeHandFunc: NoHandShakeFunc,
	}
	newTCPTransport := NewTCPTransport(config)
	assert.Equal(t, newTCPTransport.ListenAddress, "localhost:8080")
	assert.Nil(t, newTCPTransport.ListenAndAccept())
}