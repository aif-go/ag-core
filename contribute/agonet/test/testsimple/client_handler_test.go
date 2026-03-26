package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"ag-core/contribute/agonet/simple/codec"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestClientHandler(t *testing.T) {
	addr := "localhost:9000"

	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6061", nil)
	}()

	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		hexStr := hex.EncodeToString(msg)
		fmt.Printf("Received msg: %s, len: %d, hexStr: %s\n", string(msg), len(msg), hexStr)
		fmt.Printf("client Received reply message: %s\n", msg)
		ctx.FireRead(msg)
	})

	lengthDecod := codec.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	lengthEncod := codec.NewLengthFieldEncoder(nil, 2, 0, false)
	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
				lengthDecod,
				lengthEncod,
				testHand,
			)

		// c.Pipeline().AddLast(testHand)
		return nil
	}

	evenhand, err := simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(pipelineInitializer),
	)
	if err != nil {
		t.Fatalf("NewSimpleEventHandlerWithOptions failed: %v", err)
	}

	client, err := agonet.NewClient(evenhand, &agonet.ClientConfig{})
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	tcon, err := client.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial failed: %v", err)
	}

	channel, err := simple.ChannelForConn(tcon)
	if err != nil {
		t.Fatalf("ChannelForConn failed: %v", err)
	}
	// // 获取tcon 的context
	// tmpCtx := tcon.Context()
	// channel, ok := tmpCtx.(simple.Channel)
	// if !ok {
	// 	t.Fatalf("tmpCtx is not simple.Channel")
	// }

	channel.Write("abc")
	channel.Write("sirius")
	channel.Write("张三")
	channel.Write("hello world")

	channel.Close(nil)

	channel.IsActive()

	time.Sleep(time.Second)
	// tcon.W

}
