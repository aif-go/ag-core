package test

import (
	"ag-core/contribute/agonet/pkg/pool/byteslice"
	"fmt"
	"testing"

	"github.com/smallnest/ringbuffer"
)

func TestRingBuffer(t *testing.T) {

	// inboundBytes := byteslice.Get(1024 * 1024 * 1)
	inboundBytes := byteslice.Get(1)
	// ringBuffer := ringbuffer.New(1024) // TODO 环形缓冲区大小
	ringBuffer := ringbuffer.NewBuffer(inboundBytes)
	ringBuffer.SetBlocking(true)

	msg := []byte("helloworld")
	n, err := ringBuffer.Write(msg)

	fmt.Printf("n: %d, err: %v\n", n, err)

}
