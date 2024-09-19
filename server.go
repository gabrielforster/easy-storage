package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/gabrielforster/easy-storage/protocol"
)

type FileServerOptions struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         protocol.Transport
	BootstrapNodes    []string
}

type FileServer struct {
	Options FileServerOptions

	peerLock sync.Mutex
	peers    map[string]protocol.Peer

	storage *Storage
	quitch  chan struct{}
}

func NewFileServer(options FileServerOptions) *FileServer {
	storageOptions := StorageOptions{
		Root:              options.StorageRoot,
		PathTransformFunc: options.PathTransformFunc,
	}

	return &FileServer{
		Options: options,
		storage: NewStorage(storageOptions),
		quitch:  make(chan struct{}),
		peers:   make(map[string]protocol.Peer),
	}
}

type Payload struct {
	Key string
	// todo change this for io.Reader so we can stream
	// instead of copy all file data once
	Data []byte
}

func (s *FileServer) broadcast(payload *Payload) error {
	peers := []io.Writer{}
	for _, node := range s.peers {
		peers = append(peers, node)
	}

	// this is possible because the Peer interface "extends" net.Conn
	// and net.Conn implements whe Writer and Reader interfaces.
	mw := io.MultiWriter(peers...)
	return gob.NewEncoder(mw).Encode(payload)
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	buf := new(bytes.Buffer)
	tee := io.TeeReader(r, buf)

	if err := s.storage.Write(key, tee); err != nil {
		return err
	}

	payload := &Payload{
		Key:  key,
		Data: buf.Bytes(),
	}

  return s.broadcast(payload)
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) OnPeer(peer protocol.Peer) error {
	s.peerLock.Lock()
	defer s.peerLock.Unlock()
	s.peers[peer.RemoteAddr().String()] = peer

	log.Printf("connected with remote: %s\n", peer.RemoteAddr())

	return nil
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to error or user quit action")
		s.Options.Transport.Close()
	}()

	for {
		select {
		case msg := <-s.Options.Transport.Consume():
			var p Payload
      // fixme: eof error on decoding the payload
			if err := gob.NewDecoder(bytes.NewReader(msg.Payload)).Decode(&p); err != nil {
        log.Printf("error while decoding incoming message: %+v\n", err)
			}
      fmt.Printf("received: %+v\n", p)
		case <-s.quitch:
			return
		}
	}
}
func (s *FileServer) bootstrapNetwork() error {
	for _, addr := range s.Options.BootstrapNodes {
		if len(addr) == 0 {
			continue
		}

		go func(addr string) {
			fmt.Printf("[%s] attemping to connect with remote %s\n", s.Options.Transport.Addr(), addr)
			if err := s.Options.Transport.Dial(addr); err != nil {
				log.Println("dial error: ", err)
			}
		}(addr)
	}
	return nil
}

func (s *FileServer) Start() error {
	if err := s.Options.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()
	s.loop()

	return nil
}
