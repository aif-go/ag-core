package testsimple

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"fmt"
	"net/http"
	"sync"
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
		// hexStr := hex.EncodeToString(msg)
		// fmt.Printf("Received msg: %s, len: %d, hexStr: %s\n", string(msg), len(msg), hexStr)
		fmt.Printf("client Received reply message: %s\n", msg)
		ctx.FireRead(msg)
	})

	// lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	// lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)
	// 通道激活事件
	activeHand := simple.ActiveHandlerFunc(func(ctx simple.ActiveContext) {
		fmt.Printf("test active, remote addr: %s\n", ctx.Channel().RemoteAddr())
	})

	// 通道非激活事件
	inactiveHand := simple.InactiveHandlerFunc(func(ctx simple.InactiveContext, ex error) {
		fmt.Printf("test inactive, remote addr: %s, reason: %v\n", ctx.Channel().RemoteAddr(), ex)
	})
	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
				activeHand,
				inactiveHand,
				// lengthDecod,
				// lengthEncod,
				simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2),
				simple.NewLengthFieldEncoder(nil, 2, 0, false),
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

	var do = func() {
		tcon, err := client.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Dial failed: %v", err)
		}

		channel, err := simple.ChannelFromConn(tcon)
		if err != nil {
			t.Fatalf("ChannelForConn failed: %v", err)
		}
		// // 获取tcon 的context
		// tmpCtx := tcon.Context()
		// channel, ok := tmpCtx.(simple.Channel)
		// if !ok {
		// 	t.Fatalf("tmpCtx is not simple.Channel")
		// }

		channel.Write("12345")
		channel.Write("123456789")
		// channel.Write("abc")
		// channel.Write("sirius")
		// channel.Write("张三")
		// channel.Write("hello world")
		// channel.Write("1")
		// channel.Write("2")
		// channel.Write("3")
		// channel.Write("4")
		// channel.Write("5")

		time.Sleep(time.Second * 60)
		// time.Sleep(time.Millisecond * 1)
		channel.Close(nil)

		channel.IsActive()
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		// go func() {
		func() {
			// for {
			do()
			// }
			defer wg.Done()
		}()
	}

	wg.Wait()

}
