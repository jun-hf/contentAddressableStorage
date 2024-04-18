package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, any) error
}

type GOBDecoder struct {}

func (g GOBDecoder) Decode(c io.Reader, msg any) error {
	return gob.NewDecoder(c).Decode(msg)
}