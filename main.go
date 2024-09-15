package main

import (
	"fmt"
	"log"

	"github.com/gabrielforster/easy-storage/protocol"
)

func main() {
	options := protocol.TCPTransportOptions{
		ListenAddr:    ":3000",
		HandShakeFunc: protocol.DumbHandShakeFunc,
		Decoder:       protocol.DefaultDecoder{},
		OnPeer:        HandlePeer,
	}

	transport := protocol.NewTCPTranport(options)

	go func() {
		for {
			msg := <-transport.Consume()
			fmt.Printf("new message: %+v\n", msg)
			payloadAsString := string(msg.Payload)
			fmt.Printf("string from message.Payload: %+v\n", payloadAsString)
		}
	}()

	if err := transport.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
func HandlePeer(peer protocol.Peer) error {
  fmt.Println("logic on OnPeer outside tcp transport")
	return nil
}
