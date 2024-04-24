package p2p

var (
	IncomingMessage = byte(1)
	IncomingStream = byte(2)
)
type Message struct {
	Payload []byte
	From string
	Stream bool
}