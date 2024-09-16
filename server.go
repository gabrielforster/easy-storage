package main

import (
	"fmt"
	"log"

	"github.com/gabrielforster/easy-storage/protocol"
)

type FileServerOptions struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         protocol.TCPTransport
}

type FileServer struct {
	Options FileServerOptions

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
	}
}

func (s *FileServer) Stop() {
	close(s.quitch)
}

func (s *FileServer) loop() {
	defer func() {
    log.Println("file server stopped due to user quit action")
    s.Options.Transport.Close()
  }()

	for {
		select {
		case msg := <-s.Options.Transport.Consume():
			fmt.Println(msg)
		case <-s.quitch:
			return
		}
	}
}

func (s *FileServer) Start() error {
	if err := s.Options.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.loop()

	return nil
}
