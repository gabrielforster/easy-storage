package protocol

import "net"

// Peer is the interface representation of a remote node.
type Peer interface {
	net.Conn
	Send([]byte) error
	CloseStream()
}

// Transport is anything that handles the communication
// between the nodes in the network. This can be TCP, UDP, websockets, ...
type Transport interface {
	Addr() string
	Dial(string) error
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
