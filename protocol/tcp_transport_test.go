package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOptions{
		ListenAddr:    ":3000",
		HandShakeFunc: DumbHandShakeFunc,
		Decoder:       DefaultDecoder{},
	}

	tr := NewTCPTranport(opts)
	assert.Equal(t, tr.Options.ListenAddr, opts.ListenAddr)

	assert.Nil(t, tr.ListenAndAccept())
}
