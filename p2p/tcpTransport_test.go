package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)


func TestTCPTransport(t *testing.T) {
	tcpAddress := ":8080"
	tcpT := NewTCPTransport(tcpAddress)
	assert.Equal(t, tcpT.listenAddress, tcpAddress)
	assert.Nil(t, tcpT.ListenAndAccept())
}