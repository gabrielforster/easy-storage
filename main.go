package main

import (
	"bytes"
	"log"
	"time"

	"github.com/gabrielforster/easy-storage/protocol"
)

func genServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOptions := protocol.TCPTransportOptions{
		ListenAddr:    listenAddr,
		HandShakeFunc: protocol.DumbHandShakeFunc,
		Decoder:       protocol.DefaultDecoder{},
	}
	tcpTransport := protocol.NewTCPTranport(tcpTransportOptions)

	fileServerOptions := FileServerOptions{
		StorageRoot:       listenAddr + "_root_folder",
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	server := NewFileServer(fileServerOptions)

  tcpTransport.OnPeer = server.OnPeer

	return server
}
func main() {
	s1 := genServer(":3000")
	s2 := genServer(":4000", ":3000")

	time.Sleep(1 * time.Second)

	go func() { log.Fatal(s1.Start()) }()

	time.Sleep(1 * time.Second)

	go s2.Start()

	time.Sleep(2 * time.Second)

	data := bytes.NewReader([]byte("my big data here"))
	s2.StoreData("key", data)

	select {}
}
