package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface {
	Decode(io.Reader, *Message) error
}

type GOBDecoder struct {}

func (g GOBDecoder) Decode(c io.Reader, msg *Message) error {
	return gob.NewDecoder(c).Decode(msg)
}

type DefaultDecoder struct {}

func (d DefaultDecoder) Decode(c io.Reader, msg *Message) error {
	byt := make([]byte, 1028)
	n, err := c.Read(byt)
	if err != nil {
		return err
	}
	msg.Payload = byt[:n]
	return nil
}