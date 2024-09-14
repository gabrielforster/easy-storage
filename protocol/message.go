package protocol

const (
	IncomingMessage = 0x1
	IncomingStream  = 0x2
	CloseConn       = 0xFF
)

// RPC holds any arbitrary data that is being sent between
// two nodes in the network over a transport.
type RPC struct {
	From    string
	Payload []byte
	Stream  bool
}
