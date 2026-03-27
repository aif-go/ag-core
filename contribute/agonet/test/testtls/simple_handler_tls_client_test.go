package testtls

import (
	"ag-core/contribute/agonet"
	"ag-core/contribute/agonet/simple"
	"encoding/hex"
	"fmt"
	"net/http"
	"testing"
	"time"
)

var simpleChannel simple.Channel

func TestSimpleTlsClientHandler(t *testing.T) {

	addr := "localhost:8443"

	// 启动 http pprof
	go func() {
		http.ListenAndServe(":6061", nil)
	}()

	binder, err := buildCfgBinder(_agonet_tls_cli)
	if err != nil {
		t.Fatalf("buildCfgBinder failed: %v", err)
	}

	cliCfg, err := agonet.NewClientConfig(binder)
	if err != nil {
		t.Fatalf("NewClientConfig failed: %v", err)
	}

	eventhandler, err := _simpleClientEventHandler()
	if err != nil {
		t.Fatalf("LoadAuthCert failed: %v", err)
	}

	client, err := agonet.NewClient(eventhandler, cliCfg)
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

	fmt.Printf("tcon: %s\n", tcon.RemoteAddr())
	// channel, err := simple.ChannelForConn(tcon)
	// if err != nil {
	// 	t.Fatalf("ChannelForConn failed: %v", err)
	// }

	channel := simpleChannel

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

	// tcon.W
	// channel.Close(nil)

	channel.IsActive()

	time.Sleep(time.Second)
	// channel.Close(nil)
	channel.Close(fmt.Errorf("我自己要关的"))

	// time.Sleep(time.Second) // 等待重连

	client.Stop()

}

func _simpleClientEventHandler() (agonet.EventHandler, error) {

	var testHand = simple.NewSimpleInboundHandler(func(ctx simple.InboundContext, msg []byte) {
		hexStr := hex.EncodeToString(msg)
		fmt.Printf("Received msg: %s, len: %d, hexStr: %s\n", string(msg), len(msg), hexStr)
		fmt.Printf("client Received reply message: %s\n", msg)
		ctx.FireRead(msg)
	})

	lengthDecod := simple.NewLengthFieldDecoder(nil, 1024, 0, 2, 0, 2)
	lengthEncod := simple.NewLengthFieldEncoder(nil, 2, 0, false)

	// 通道激活事件
	activeHand := simple.ActiveHandlerFunc(func(ctx simple.ActiveContext) {
		fmt.Printf("test active, remote addr: %s\n", ctx.Channel().RemoteAddr())
		simpleChannel = ctx.Channel() // 保存channel
	})

	// 通道非激活事件
	inactiveHand := simple.InactiveHandlerFunc(func(ctx simple.InactiveContext, ex error) {
		fmt.Printf("test inactive, remote addr: %s, reason: %v\n", ctx.Channel().RemoteAddr(), ex)
		// TODO 要重连怎么办
	})

	pipelineInitializer := func(c simple.Channel) error {
		c.Pipeline().
			// AddLast(&echoHandler{}).
			AddLast(
				lengthDecod,
				lengthEncod,
				testHand,
				activeHand,
				inactiveHand,
			)

		// c.Pipeline().AddLast(testHand)
		return nil
	}

	evenhand, err := simple.NewSimpleEventHandlerWithOptions(
		simple.WithChannelInitializer(pipelineInitializer),
	)
	if err != nil {
		return nil, err
	}

	return evenhand, nil
}
