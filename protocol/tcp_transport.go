package protocol

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer is the network remote node over a TCP transport
// estabilished connection
type TCPPeer struct {
	// The peer connection it self. TCP in this case
  net.Conn

	// if accepted and retrieved => true
	// if dialed and retrieved => false
	outbound bool

	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream() { p.wg.Done() }

func (p *TCPPeer) Send(data []byte) error {
	_, err := p.Conn.Write(data)
	return err
}

type TCPTransportOptions struct {
	ListenAddr    string
	HandShakeFunc HandShakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	Options    TCPTransportOptions
	listener   net.Listener
	rpcChannel chan RPC
}

func NewTCPTranport(options TCPTransportOptions) *TCPTransport {
	return &TCPTransport{
		Options:    options,
		rpcChannel: make(chan RPC, 1024),
	}
}

// Addr implements the Transport interface
// returns the address.
// transport is accepting new connections.
func (t *TCPTransport) Addr() string {
	return t.Options.ListenAddr
}

// Consume implements the Transport interface, returns a read-only
// of incoming messages received from a connected peer in the network.
func (t *TCPTransport) Consume() <-chan RPC {
	return t.rpcChannel
}

// Close implements the Transport interface
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)
	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.Options.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port: %s\n", t.Options.ListenAddr)
	return nil
}

func (t *TCPTransport) handleConn(conn net.Conn, isOutbound bool) {
	var err error

	defer func() {
		fmt.Printf("dropping peer conn: %+v", err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, isOutbound)

	if err = t.Options.HandShakeFunc(peer); err != nil {
		return
	}

	if t.Options.OnPeer != nil {
		if err = t.Options.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for {
		rpc := RPC{}
		err = t.Options.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] incoming stream, waiting...\n", conn.RemoteAddr())
			peer.wg.Wait()
			fmt.Printf("[%s] stream closed, resuming read loop\n", conn.RemoteAddr())
			continue
		}

		t.rpcChannel <- rpc
	}
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}

		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}

		go t.handleConn(conn, false)
	}
}
