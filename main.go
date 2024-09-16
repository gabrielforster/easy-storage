package main

import (
	"log"
	"time"

	"github.com/gabrielforster/easy-storage/protocol"
)

func main() {
	tcpTransportOptions := protocol.TCPTransportOptions{
		ListenAddr:    ":3000",
		HandShakeFunc: protocol.DumbHandShakeFunc,
		Decoder:       protocol.DefaultDecoder{},
		// todo add OnPeer function
	}
	tcpTransport := protocol.NewTCPTranport(tcpTransportOptions)

	fileServerOptions := FileServerOptions{
		StorageRoot:       "root_folder",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         *tcpTransport,
	}

	server := NewFileServer(fileServerOptions)
	go func() {
    time.Sleep(time.Second * 3)
    server.Stop()
	}()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
